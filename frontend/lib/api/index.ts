// API客户端统一导出
import { getAuth } from './orval/auth';
import { getMailLog } from './orval/mail-log';
import { getSmtpCredentials } from './orval/smtp-credentials';
import { getUsermgmt } from './orval/usermgmt';
import { getSystem } from './orval/system';
import { getStatus } from './orval/status';

// 类型统一导出
export * from './orval/smtpApi.schemas';
export * from './orval/auth';
export * from './orval/mail-log';
export * from './orval/smtp-credentials';
export * from './orval/usermgmt';
export * from './orval/system';
export * from './orval/status';

// API实例分组
export const authApi = (() => {
  const api = getAuth();
  return {
    login: api.postApiV1AuthLogin,
    register: api.postApiV1AuthRegister,
  };
})();

const smtpCredentialsApi = getSmtpCredentials();

export const credentialsApi = {
  list: smtpCredentialsApi.getApiV1Credentials,
  create: smtpCredentialsApi.postApiV1Credentials,
  get: smtpCredentialsApi.getApiV1CredentialsId,
  update: smtpCredentialsApi.putApiV1CredentialsId,
  deleteCredential: smtpCredentialsApi.deleteApiV1CredentialsId,
  resetPassword: smtpCredentialsApi.postApiV1CredentialsIdResetPassword,
};

export const logsApi = (() => {
  const api = getMailLog();
  return {
    list: api.getApiV1Logs,
    get: api.getApiV1LogsId,
    recent: api.getApiV1LogsRecent,
  };
})();

export const userApi = (() => {
  const api = getUsermgmt();
  return {
    profile: api.getApiV1User,
    update: api.putApiV1User,
  };
})();

export const systemApi = (() => {
  const api = getSystem();
  return {
    health: api.getHealth,
  };
})();

export const statsApi = (() => {
  const api = getStatus();
  return {
    overview: api.getApiV1Stats,
    quota: api.getApiV1StatsQuota,
  };
})(); 