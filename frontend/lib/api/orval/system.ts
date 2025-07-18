/**
 * Generated by orval v6.31.0 🍺
 * Do not edit manually.
 * SMTP中继服务API
 * 基于Go+Gin的SMTP转发中继服务API文档
 * OpenAPI spec version: 1.0
 */
import type { ApiRelayInfoResponse, GetHealth200 } from "./smtpApi.schemas";
import { customAxios } from "../custom-axios";

export const getSystem = () => {
  /**
   * 获取SMTP中继域名和IP
   * @summary 获取SMTP中继信息
   */
  const getApiRelayInfo = () => {
    return customAxios<ApiRelayInfoResponse>({
      url: `/api/relay-info`,
      method: "GET",
    });
  };
  /**
   * 检查服务健康状态
   * @summary 健康检查
   */
  const getHealth = () => {
    return customAxios<GetHealth200>({ url: `/health`, method: "GET" });
  };
  return { getApiRelayInfo, getHealth };
};
export type GetApiRelayInfoResult = NonNullable<
  Awaited<ReturnType<ReturnType<typeof getSystem>["getApiRelayInfo"]>>
>;
export type GetHealthResult = NonNullable<
  Awaited<ReturnType<ReturnType<typeof getSystem>["getHealth"]>>
>;
