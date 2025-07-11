// API错误响应类型
export interface ApiErrorResponse {
  message: string;
  status?: number;
  code?: string;
  details?: any;
}

// 扩展axios配置类型，添加metadata
declare module 'axios' {
  interface InternalAxiosRequestConfig {
    metadata?: {
      requestStartTime: number;
    };
  }
}

// 认证相关类型
export interface AuthUser {
  id: string;
  username: string;
  email?: string;
  settings?: UserSettings;
  created_at: string;
  updated_at: string;
}

export interface UserSettings {
  allowed_domains?: string[];
  daily_quota?: number;
  hourly_quota?: number;
}

// 登录请求和响应类型
export interface LoginRequest {
  username: string;
  password: string;
  remember_me?: boolean;
}

export interface LoginResponse {
  success: boolean;
  token: string;
  user: AuthUser;
  expires_at: string;
}

// SMTP凭据类型
export interface SMTPCredential {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  username: string;
  password: string;
  settings: SMTPCredentialSettings;
  created_at: string;
  updated_at: string;
}

export interface SMTPCredentialSettings {
  allowed_domains?: string[];
  daily_quota?: number;
  hourly_quota?: number;
  max_recipients?: number;
}

// MailLog类型
export interface MailLog {
  id: string;
  user_id: string;
  credential_id: string;
  from: string;
  to: string[];
  cc?: string[];
  bcc?: string[];
  subject: string;
  status: MailStatus;
  error_message?: string;
  sent_at?: string;
  created_at: string;
  updated_at: string;
}

export type MailStatus = 'queued' | 'sending' | 'sent' | 'failed';

// 统计信息类型
export interface MailStats {
  total_sent: number;
  success_rate: number;
  daily_sent: number;
  hourly_sent: number;
  failed_count: number;
  queue_count: number;
}

export interface QuotaStats {
  daily_quota: number;
  daily_used: number;
  hourly_quota: number;
  hourly_used: number;
  remaining_daily: number;
  remaining_hourly: number;
}

// API分页响应类型
export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// API通用响应类型
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
} 