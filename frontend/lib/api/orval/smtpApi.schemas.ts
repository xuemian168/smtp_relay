/**
 * Generated by orval v6.31.0 🍺
 * Do not edit manually.
 * SMTP中继服务API
 * 基于Go+Gin的SMTP转发中继服务API文档
 * OpenAPI spec version: 1.0
 */
export type GetHealth200 = { [key: string]: unknown };

export type GetApiV1LogsRecentStatus =
  (typeof GetApiV1LogsRecentStatus)[keyof typeof GetApiV1LogsRecentStatus];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const GetApiV1LogsRecentStatus = {
  queued: "queued",
  sending: "sending",
  sent: "sent",
  failed: "failed",
} as const;

export type GetApiV1LogsRecentParams = {
  /**
   * 最近N天
   */
  days?: number;
  /**
   * 邮件状态
   */
  status?: GetApiV1LogsRecentStatus;
  /**
   * 页码
   */
  page?: number;
  /**
   * 每页数量
   */
  page_size?: number;
};

export type GetApiV1LogsStatus =
  (typeof GetApiV1LogsStatus)[keyof typeof GetApiV1LogsStatus];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const GetApiV1LogsStatus = {
  queued: "queued",
  sending: "sending",
  sent: "sent",
  failed: "failed",
} as const;

export type GetApiV1LogsParams = {
  /**
   * 页码
   */
  page?: number;
  /**
   * 每页数量
   */
  page_size?: number;
  /**
   * 邮件状态
   */
  status?: GetApiV1LogsStatus;
  /**
   * 发件人筛选
   */
  from?: string;
  /**
   * 收件人筛选
   */
  to?: string;
  /**
   * 开始日期
   */
  date_from?: string;
  /**
   * 结束日期
   */
  date_to?: string;
};

export interface ModelsUserSettings {
  allowed_domains?: string[];
  daily_quota?: number;
  hourly_quota?: number;
}

export interface ModelsSMTPCredentialSettings {
  /** 允许发送的域名 */
  allowed_domains?: string[];
  /** 该凭据的日配额 */
  daily_quota?: number;
  /** 该凭据的小时配额 */
  hourly_quota?: number;
  /** 单封邮件最大收件人数 */
  max_recipients?: number;
}

export interface ModelsSMTPCredential {
  created_at?: string;
  /** 描述信息 */
  description?: string;
  id?: string;
  last_used?: string;
  /** 凭据名称，如"mailcow-server1" */
  name?: string;
  settings?: ModelsSMTPCredentialSettings;
  /** active, disabled */
  status?: string;
  updated_at?: string;
  /** 使用次数 */
  usage_count?: number;
  user_id?: string;
  /** SMTP用户名 */
  username?: string;
}

export interface ModelsMailLog {
  attempts?: number;
  completed_at?: string;
  created_at?: string;
  credential_id?: string;
  error_message?: string;
  from?: string;
  id?: string;
  last_attempt?: string;
  message_id?: string;
  relay_ip?: string;
  size?: number;
  /** queued, sending, sent, failed */
  status?: string;
  subject?: string;
  to?: string[];
  user_id?: string;
}

export interface ModelsDNSRecord {
  /** 记录名称，如 selector._domainkey.example.com */
  name?: string;
  /** 优先级（对TXT记录通常为0） */
  priority?: number;
  /** TTL值 */
  ttl?: number;
  /** TXT */
  type?: string;
  /** 记录值 */
  value?: string;
}

export interface ModelsDKIMValidationResult {
  checked_at?: string;
  dns_found?: boolean;
  dns_record?: string;
  domain?: string;
  error_message?: string;
  expected_dns?: string;
  selector?: string;
  valid?: boolean;
}

export interface ModelsDKIMKeyPair {
  /** 签名算法（rsa-sha256） */
  algorithm?: string;
  created_at?: string;
  /** 生成的DNS TXT记录 */
  dns_record?: string;
  /** DNS记录是否已验证 */
  dns_verified?: boolean;
  /** 签名域名 */
  domain?: string;
  /** 密钥过期时间 */
  expires_at?: string;
  id?: string;
  /** 密钥长度（1024, 2048） */
  key_size?: number;
  /** 最后验证时间 */
  last_verified?: string;
  /** 公钥 */
  public_key?: string;
  /** DKIM选择器 */
  selector?: string;
  /** active, inactive, expired */
  status?: string;
  updated_at?: string;
  user_id?: string;
}

export interface ApiUserInfoResponse {
  created_at?: string;
  email?: string;
  id?: string;
  settings?: ModelsUserSettings;
  status?: string;
  username?: string;
}

export interface ApiUserInfo {
  email?: string;
  id?: string;
  username?: string;
}

export interface ApiUpdateUserInfoRequest {
  settings?: ModelsUserSettings;
  /**
   * @minLength 3
   * @maxLength 50
   */
  username?: string;
}

export interface ApiUpdateCredentialRequest {
  /** @maxLength 200 */
  description?: string;
  /**
   * @minLength 1
   * @maxLength 50
   */
  name: string;
  settings?: ModelsSMTPCredentialSettings;
}

export interface ApiStatsResponse {
  data?: ApiStatsResponseData;
  success?: boolean;
}

export type ApiStatsResponseDataMailStats = { [key: string]: unknown };

export type ApiStatsResponseDataCredentialStatsItem = {
  [key: string]: unknown;
};

export type ApiStatsResponseData = {
  credential_count?: number;
  credential_stats?: ApiStatsResponseDataCredentialStatsItem[];
  mail_stats?: ApiStatsResponseDataMailStats;
};

export interface ApiResetPasswordResponse {
  message?: string;
  password?: string;
  success?: boolean;
}

/**
 * SMTP中继域名和IP
 */
export interface ApiRelayInfoData {
  relayDomain?: string;
  relayIP?: string;
}

export interface ApiRelayInfoResponse {
  data?: ApiRelayInfoData;
  success?: boolean;
}

export interface ApiRegisterResponse {
  success?: boolean;
  token?: string;
  user?: ApiUserInfo;
}

export interface ApiRegisterRequest {
  email: string;
  /** @minLength 8 */
  password: string;
  /**
   * @minLength 3
   * @maxLength 50
   */
  username: string;
}

export type ApiRecentMailLogsResponseDataStatistics = { [key: string]: number };

export type ApiRecentMailLogsResponseData = {
  days?: number;
  mail_logs?: ModelsMailLog[];
  page?: number;
  page_size?: number;
  pages?: number;
  statistics?: ApiRecentMailLogsResponseDataStatistics;
  total?: number;
};

export interface ApiRecentMailLogsResponse {
  data?: ApiRecentMailLogsResponseData;
  success?: boolean;
}

export type ApiQuotaStatsResponseDataUserSettings = {
  daily_quota?: number;
  hourly_quota?: number;
};

export type ApiQuotaStatsResponseDataCredentialQuotasItem = {
  [key: string]: unknown;
};

export type ApiQuotaStatsResponseData = {
  credential_quotas?: ApiQuotaStatsResponseDataCredentialQuotasItem[];
  user_settings?: ApiQuotaStatsResponseDataUserSettings;
};

export interface ApiQuotaStatsResponse {
  data?: ApiQuotaStatsResponseData;
  success?: boolean;
}

export interface ApiMailLogResponse {
  data?: ModelsMailLog;
  success?: boolean;
}

export type ApiMailLogListResponseData = {
  mail_logs?: ModelsMailLog[];
  page?: number;
  page_size?: number;
  pages?: number;
  total?: number;
};

export interface ApiMailLogListResponse {
  data?: ApiMailLogListResponseData;
  success?: boolean;
}

export interface ApiLoginResponse {
  success?: boolean;
  token?: string;
  user?: ApiUserInfo;
}

export interface ApiLoginRequest {
  password: string;
  username: string;
}

export interface ApiDNSRecordResponse {
  data?: ModelsDNSRecord;
  success?: boolean;
}

export interface ApiDKIMValidationResponse {
  data?: ModelsDKIMValidationResult;
  success?: boolean;
}

export interface ApiDKIMKeyPairResponse {
  data?: ModelsDKIMKeyPair;
  success?: boolean;
}

export interface ApiDKIMKeyPairListResponse {
  data?: ModelsDKIMKeyPair[];
  success?: boolean;
}

export interface ApiCredentialResponse {
  data?: ModelsSMTPCredential;
  success?: boolean;
}

export interface ApiCredentialListResponse {
  data?: ModelsSMTPCredential[];
  success?: boolean;
}

export interface ApiCreateDKIMKeyRequest {
  domain: string;
  key_size?: number;
  selector: string;
}

export type ApiCreateCredentialResponseData = {
  credential?: ModelsSMTPCredential;
  password?: string;
};

export interface ApiCreateCredentialResponse {
  data?: ApiCreateCredentialResponseData;
  success?: boolean;
}

export interface ApiCreateCredentialRequest {
  /** @maxLength 200 */
  description?: string;
  /**
   * @minLength 1
   * @maxLength 50
   */
  name: string;
}

export interface ApiAPIResponse {
  data?: unknown;
  error?: string;
  message?: string;
  success?: boolean;
}
