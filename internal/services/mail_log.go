package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"smtp-relay/internal/database"
	"smtp-relay/internal/models"
)

// MailLogService MailLog管理服务
type MailLogService struct {
	db     *database.MongoDB
	logger *logrus.Logger
}

// NewMailLogService 创建MailLog管理服务
func NewMailLogService(db *database.MongoDB, logger *logrus.Logger) *MailLogService {
	return &MailLogService{
		db:     db,
		logger: logger,
	}
}

// GetMailLogsByUser 获取用户的MailLog列表（支持分页和过滤）
func (s *MailLogService) GetMailLogsByUser(userID primitive.ObjectID, page, pageSize int, status, from, to, dateFrom, dateTo string) ([]*models.MailLog, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("mail_logs")

	// 构建过滤条件
	filter := bson.M{"user_id": userID}

	// 状态过滤
	if status != "" {
		filter["status"] = status
	}

	// 发件人过滤
	if from != "" {
		filter["from"] = bson.M{"$regex": from, "$options": "i"}
	}

	// 收件人过滤
	if to != "" {
		filter["to"] = bson.M{"$elemMatch": bson.M{"$regex": to, "$options": "i"}}
	}

	// 日期范围过滤
	if dateFrom != "" || dateTo != "" {
		dateFilter := bson.M{}
		if dateFrom != "" {
			if fromTime, err := time.Parse("2006-01-02", dateFrom); err == nil {
				dateFilter["$gte"] = fromTime
			}
		}
		if dateTo != "" {
			if toTime, err := time.Parse("2006-01-02", dateTo); err == nil {
				dateFilter["$lte"] = toTime.Add(24 * time.Hour)
			}
		}
		if len(dateFilter) > 0 {
			filter["created_at"] = dateFilter
		}
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 计算分页
	skip := (page - 1) * pageSize
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSort(bson.D{{"created_at", -1}}) // 按创建时间倒序

	// 查询数据
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var mailLogs []*models.MailLog
	for cursor.Next(ctx) {
		var mailLog models.MailLog
		if err := cursor.Decode(&mailLog); err != nil {
			continue
		}
		mailLogs = append(mailLogs, &mailLog)
	}

	return mailLogs, total, nil
}

// GetMailLogByID 获取单个MailLog
func (s *MailLogService) GetMailLogByID(userID, mailLogID primitive.ObjectID) (*models.MailLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("mail_logs")
	filter := bson.M{
		"_id":     mailLogID,
		"user_id": userID,
	}

	var mailLog models.MailLog
	err := collection.FindOne(ctx, filter).Decode(&mailLog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &mailLog, nil
}

// GetMailLogsByCredential 获取指定凭据的MailLog
func (s *MailLogService) GetMailLogsByCredential(userID, credentialID primitive.ObjectID, page, pageSize int) ([]*models.MailLog, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("mail_logs")
	filter := bson.M{
		"user_id":       userID,
		"credential_id": credentialID,
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 计算分页
	skip := (page - 1) * pageSize
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSort(bson.D{{"created_at", -1}})

	// 查询数据
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var mailLogs []*models.MailLog
	for cursor.Next(ctx) {
		var mailLog models.MailLog
		if err := cursor.Decode(&mailLog); err != nil {
			continue
		}
		mailLogs = append(mailLogs, &mailLog)
	}

	return mailLogs, total, nil
}

// GetUserMailStats 获取用户邮件统计信息
func (s *MailLogService) GetUserMailStats(userID primitive.ObjectID) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("mail_logs")

	// 获取今日统计
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	todayFilter := bson.M{
		"user_id": userID,
		"created_at": bson.M{
			"$gte": today,
			"$lt":  tomorrow,
		},
	}

	todayTotal, _ := collection.CountDocuments(ctx, todayFilter)

	// 获取本月统计
	thisMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	nextMonth := thisMonth.AddDate(0, 1, 0)

	monthFilter := bson.M{
		"user_id": userID,
		"created_at": bson.M{
			"$gte": thisMonth,
			"$lt":  nextMonth,
		},
	}

	monthTotal, _ := collection.CountDocuments(ctx, monthFilter)

	// 获取状态统计
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$group": bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	statusCounts := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		statusCounts[result.ID] = result.Count
	}

	// 计算成功率
	total := int64(0)
	sent := statusCounts["sent"]
	for _, count := range statusCounts {
		total += count
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(sent) / float64(total) * 100
	}

	return map[string]interface{}{
		"today_total":   todayTotal,
		"month_total":   monthTotal,
		"total":         total,
		"status_counts": statusCounts,
		"success_rate":  successRate,
	}, nil
}

// GetCredentialMailStats 获取凭据邮件统计信息
func (s *MailLogService) GetCredentialMailStats(userID, credentialID primitive.ObjectID) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("mail_logs")

	// 获取今日统计
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	todayFilter := bson.M{
		"user_id":       userID,
		"credential_id": credentialID,
		"created_at": bson.M{
			"$gte": today,
			"$lt":  tomorrow,
		},
	}

	todayTotal, _ := collection.CountDocuments(ctx, todayFilter)

	// 获取本小时统计
	thisHour := time.Now().Truncate(time.Hour)
	nextHour := thisHour.Add(time.Hour)

	hourFilter := bson.M{
		"user_id":       userID,
		"credential_id": credentialID,
		"created_at": bson.M{
			"$gte": thisHour,
			"$lt":  nextHour,
		},
	}

	hourTotal, _ := collection.CountDocuments(ctx, hourFilter)

	return map[string]interface{}{
		"today_total": todayTotal,
		"hour_total":  hourTotal,
	}, nil
}

// GetRecentMailLogsByUser 获取用户近期发信历史（优化版本）
func (s *MailLogService) GetRecentMailLogsByUser(userID primitive.ObjectID, days, page, pageSize int, status string) ([]*models.MailLog, int64, map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("mail_logs")

	// 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// 构建过滤条件
	filter := bson.M{
		"user_id": userID,
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	// 状态过滤
	if status != "" {
		filter["status"] = status
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, nil, err
	}

	// 计算分页
	skip := (page - 1) * pageSize
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSort(bson.D{{"created_at", -1}}) // 按创建时间倒序

	// 查询数据
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, nil, err
	}
	defer cursor.Close(ctx)

	var mailLogs []*models.MailLog
	for cursor.Next(ctx) {
		var mailLog models.MailLog
		if err := cursor.Decode(&mailLog); err != nil {
			continue
		}
		mailLogs = append(mailLogs, &mailLog)
	}

	// 获取状态统计
	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}},
	}

	statsCursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return mailLogs, total, nil, err
	}
	defer statsCursor.Close(ctx)

	statusCounts := make(map[string]int64)
	for statsCursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := statsCursor.Decode(&result); err != nil {
			continue
		}
		statusCounts[result.ID] = result.Count
	}

	// 构建统计信息
	statistics := map[string]interface{}{
		"total_sent":    statusCounts["sent"],
		"total_failed":  statusCounts["failed"],
		"total_queued":  statusCounts["queued"],
		"total_sending": statusCounts["sending"],
		"date_range": map[string]string{
			"from": startDate.Format("2006-01-02"),
			"to":   endDate.Format("2006-01-02"),
		},
	}

	return mailLogs, total, statistics, nil
}
