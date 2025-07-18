/**
 * Generated by orval v6.31.0 🍺
 * Do not edit manually.
 * SMTP中继服务API
 * 基于Go+Gin的SMTP转发中继服务API文档
 * OpenAPI spec version: 1.0
 */
import type {
  ApiMailLogListResponse,
  ApiMailLogResponse,
  ApiRecentMailLogsResponse,
  GetApiV1LogsParams,
  GetApiV1LogsRecentParams,
} from "./smtpApi.schemas";
import { customAxios } from "../custom-axios";

export const getMailLog = () => {
  /**
   * 获取当前用户的邮件发送日志，支持分页和筛选
   * @summary 获取MailLog
   */
  const getApiV1Logs = (params?: GetApiV1LogsParams) => {
    return customAxios<ApiMailLogListResponse>({
      url: `/api/v1/logs`,
      method: "GET",
      params,
    });
  };
  /**
   * 获取指定ID的MailLog详情
   * @summary 获取单个MailLog
   */
  const getApiV1LogsId = (id: string) => {
    return customAxios<ApiMailLogResponse>({
      url: `/api/v1/logs/${id}`,
      method: "GET",
    });
  };
  /**
   * 获取用户近期发信历史，支持天数筛选和状态筛选，包含统计信息
   * @summary 获取近期MailLog
   */
  const getApiV1LogsRecent = (params?: GetApiV1LogsRecentParams) => {
    return customAxios<ApiRecentMailLogsResponse>({
      url: `/api/v1/logs/recent`,
      method: "GET",
      params,
    });
  };
  return { getApiV1Logs, getApiV1LogsId, getApiV1LogsRecent };
};
export type GetApiV1LogsResult = NonNullable<
  Awaited<ReturnType<ReturnType<typeof getMailLog>["getApiV1Logs"]>>
>;
export type GetApiV1LogsIdResult = NonNullable<
  Awaited<ReturnType<ReturnType<typeof getMailLog>["getApiV1LogsId"]>>
>;
export type GetApiV1LogsRecentResult = NonNullable<
  Awaited<ReturnType<ReturnType<typeof getMailLog>["getApiV1LogsRecent"]>>
>;
