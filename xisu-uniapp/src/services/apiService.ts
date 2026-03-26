/**
 * API 服务 - 封装网络请求
 */

// 后端API地址 - Go 后端
const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://api.xisu.leisureea.cn/api/v1';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyObject = Record<string, any>;

interface RequestOptions {
  url: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  data?: string | AnyObject | ArrayBuffer;
  header?: Record<string, string>;
}

interface ApiResponse<T = unknown> {
  code?: number;
  data?: T;
  message?: string;
  timestamp?: string;
}

// 判断响应是否为包装格式
function isWrappedResponse<T>(response: unknown): response is ApiResponse<T> {
  return (
    typeof response === 'object' &&
    response !== null &&
    'code' in response &&
    'data' in response
  );
}

interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

const IS_DEV = import.meta.env.DEV;

const debugLog = (...args: unknown[]) => {
  if (IS_DEV) {
    // eslint-disable-next-line no-console
    console.log(...args);
  }
};

const debugError = (...args: unknown[]) => {
  if (IS_DEV) {
    // eslint-disable-next-line no-console
    console.error(...args);
  }
};

/**
 * 保存登录凭证
 */
export const saveTokens = (tokens: AuthTokens) => {
  uni.setStorageSync('access_token', tokens.accessToken);
  uni.setStorageSync('refresh_token', tokens.refreshToken);
};

/**
 * 获取 Access Token
 */
export const getAccessToken = (): string | null => {
  return uni.getStorageSync('access_token') || null;
};

/**
 * 清除登录凭证
 */
export const clearTokens = () => {
  uni.removeStorageSync('access_token');
  uni.removeStorageSync('refresh_token');
  uni.removeStorageSync('user_info');
};

/**
 * 保存用户信息
 */
export const saveUserInfo = (userInfo: unknown) => {
  uni.setStorageSync('user_info', JSON.stringify(userInfo));
};

/**
 * 获取用户信息
 */
export const getUserInfo = <T = unknown>(): T | null => {
  const info = uni.getStorageSync('user_info');
  if (info) {
    try {
      return JSON.parse(info) as T;
    } catch {
      return null;
    }
  }
  return null;
};

/**
 * 获取 Refresh Token
 */
export const getRefreshToken = (): string | null => {
  return uni.getStorageSync('refresh_token') || null;
};

// Token 刷新状态管理
let isRefreshing = false;
let refreshSubscribers: Array<(token: string) => void> = [];

/**
 * 添加等待刷新的请求到队列
 */
const subscribeTokenRefresh = (callback: (token: string) => void) => {
  refreshSubscribers.push(callback);
};

/**
 * 刷新完成后，执行队列中的所有请求
 */
const onTokenRefreshed = (newToken: string) => {
  refreshSubscribers.forEach(callback => callback(newToken));
  refreshSubscribers = [];
};

/**
 * 刷新 Access Token
 */
const refreshAccessToken = async (): Promise<string | null> => {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    return null;
  }

  try {
    debugLog('[Token] Attempting to refresh access token...');
    
    return new Promise((resolve, reject) => {
      uni.request({
        url: `${BASE_URL}/auth/refresh`,
        method: 'POST',
        data: { refreshToken },
        header: { 'Content-Type': 'application/json' },
        success: (res) => {
          const statusCode = res.statusCode;
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const response = res.data as any;
          
          debugLog('[Token] Refresh response:', statusCode);
          
          if (statusCode >= 200 && statusCode < 300) {
            // 处理包装格式响应
            const data = response.code === 0 ? response.data : response;
            
            if (data?.accessToken && data?.refreshToken) {
              // 保存新的 token
              saveTokens({
                accessToken: data.accessToken,
                refreshToken: data.refreshToken,
              });
              debugLog('[Token] Token refreshed successfully');
              resolve(data.accessToken);
            } else {
              debugLog('[Token] Invalid refresh response data');
              resolve(null);
            }
          } else {
            debugLog('[Token] Refresh failed with status:', statusCode);
            resolve(null);
          }
        },
        fail: (err) => {
          debugError('[Token] Refresh request failed:', err);
          resolve(null);
        }
      });
    });
  } catch (error) {
    debugError('[Token] Refresh error:', error);
    return null;
  }
};

/**
 * 处理 Token 过期 - 尝试刷新或跳转登录
 */
const handleTokenExpired = async <T>(options: RequestOptions): Promise<T> => {
  // 如果正在刷新，等待刷新完成后重试
  if (isRefreshing) {
    return new Promise((resolve, reject) => {
      subscribeTokenRefresh(async (newToken: string) => {
        try {
          const result = await requestWithToken<T>(options, newToken);
          resolve(result);
        } catch (err) {
          reject(err);
        }
      });
    });
  }

  isRefreshing = true;

  try {
    const newAccessToken = await refreshAccessToken();
    
    if (newAccessToken) {
      // 刷新成功，通知队列中的请求
      onTokenRefreshed(newAccessToken);
      // 重试当前请求
      return await requestWithToken<T>(options, newAccessToken);
    } else {
      // 刷新失败，跳转登录
      debugLog('[Token] Refresh failed, redirecting to login...');
      clearTokens();
      uni.reLaunch({ url: '/pages/login/index' });
      throw new Error('登录已过期，请重新登录');
    }
  } finally {
    isRefreshing = false;
  }
};

/**
 * 使用指定 token 发送请求
 */
const requestWithToken = <T>(options: RequestOptions, token: string): Promise<T> => {
  const { url, method = 'GET', data, header = {} } = options;
  
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...header,
        'Authorization': `Bearer ${token}`,
      },
      success: (res) => {
        const statusCode = res.statusCode;
        const response = res.data as ApiResponse<T>;
        
        if (statusCode >= 200 && statusCode < 300) {
          if (isWrappedResponse<T>(response)) {
            if (response.code === 0) {
              resolve(response.data as T);
            } else {
              reject(new Error(response.message || '请求失败'));
            }
          } else {
            resolve(response as T);
          }
        } else {
          const errorMessage = (response as ApiResponse<T>).message || `请求失败 (${statusCode})`;
          reject(new Error(errorMessage));
        }
      },
      fail: (err) => {
        reject(new Error('网络请求失败，请检查网络连接'));
      }
    });
  });
};

/**
 * 通用请求方法 - 支持自动刷新 Token
 */
export const request = async <T = unknown>(options: RequestOptions): Promise<T> => {
  const { url, method = 'GET', data, header = {} } = options;
  
  // 获取token
  const token = getAccessToken();
  if (token) {
    header['Authorization'] = `Bearer ${token}`;
  }
  
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...header
      },
      success: async (res) => {
        const statusCode = res.statusCode;
        const response = res.data as ApiResponse<T>;
        
        debugLog('[API Response]', url, statusCode);
        
        if (statusCode >= 200 && statusCode < 300) {
          // 检查是否为包装格式 { code, data, message }
          if (isWrappedResponse<T>(response)) {
            debugLog('[API] Unwrapping response');
            if (response.code === 0) {
              resolve(response.data as T);
            } else {
              reject(new Error(response.message || '请求失败'));
            }
          } else {
            // 直接返回数据（兼容未包装的响应）
            resolve(response as T);
          }
        } else if (statusCode === 401) {
          // Token过期，尝试刷新 Token
          debugLog('[API] 401 Unauthorized, attempting token refresh...');
          
          // 排除登录和刷新接口本身，避免死循环
          if (url.includes('/auth/refresh')) {
            clearTokens();
            uni.reLaunch({ url: '/pages/login/index' });
            reject(new Error('登录已过期，请重新登录'));
            return;
          }

          // 登录接口返回401，直接提示后端错误信息
          if (url.includes('/auth/login')) {
            const errorMessage = response.message || '学号或密码错误';
            reject(new Error(errorMessage));
            return;
          }
          
          try {
            const result = await handleTokenExpired<T>(options);
            resolve(result);
          } catch (err) {
            reject(err);
          }
        } else {
          // 错误响应
          const errorMessage = response.message || `请求失败 (${statusCode})`;
          reject(new Error(errorMessage));
        }
      },
      fail: (err) => {
        debugError('Request failed:', err);
        reject(new Error('网络请求失败，请检查网络连接'));
      }
    });
  });
};

/**
 * GET 请求
 */
export const get = <T = unknown>(url: string, data?: AnyObject): Promise<T> => {
  return request<T>({ url, method: 'GET', data });
};

/**
 * POST 请求
 */
export const post = <T = unknown>(url: string, data?: AnyObject): Promise<T> => {
  return request<T>({ url, method: 'POST', data });
};

/**
 * PUT 请求
 */
export const put = <T = unknown>(url: string, data?: AnyObject): Promise<T> => {
  return request<T>({ url, method: 'PUT', data });
};

/**
 * DELETE 请求
 */
export const del = <T = unknown>(url: string, data?: AnyObject): Promise<T> => {
  return request<T>({ url, method: 'DELETE', data });
};

/**
 * 文件上传（multipart/form-data）
 */
export const uploadFile = <T = unknown>(
  url: string,
  filePath: string,
  name = 'file',
  formData?: Record<string, string>
): Promise<T> => {
  const token = getAccessToken();

  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: `${BASE_URL}${url}`,
      filePath,
      name,
      formData,
      header: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success: (res) => {
        const statusCode = res.statusCode;
        let response: unknown = res.data;

        if (typeof res.data === 'string') {
          try {
            response = JSON.parse(res.data);
          } catch {
            reject(new Error('上传响应解析失败'));
            return;
          }
        }

        if (statusCode >= 200 && statusCode < 300) {
          if (isWrappedResponse<T>(response)) {
            if (response.code === 0) {
              resolve(response.data as T);
            } else {
              reject(new Error(response.message || '上传失败'));
            }
          } else {
            resolve(response as T);
          }
          return;
        }

        const wrapped = response as ApiResponse<T>;
        reject(new Error(wrapped?.message || `上传失败 (${statusCode})`));
      },
      fail: () => {
        reject(new Error('文件上传失败，请检查网络连接'));
      },
    });
  });
};

// ============================================
// Auth API
// ============================================

export interface LoginRequest {
  studentId: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: UserInfo;
}

export interface UserInfo {
  id: string;
  username: string;
  email: string;
  studentId: string;
  realName?: string;
  avatar?: string;
  role: string;
  college?: string;
  major?: string;
  className?: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
  studentId: string;
  xiwaiPassword: string;
  emailToken?: string;
  avatar?: string;
}

export interface SendCodeRequest {
  email: string;
}

export interface VerifyCodeRequest {
  email: string;
  code: string;
}

export interface VerifyCodeResponse {
  verified: boolean;
  emailToken: string;
}

/**
 * Auth API
 */
export const authApi = {
  /** 发送邮箱验证码 */
  sendCode: (data: SendCodeRequest) => post<{ message: string }>('/auth/send-code', data),
  
  /** 验证邮箱验证码 */
  verifyCode: (data: VerifyCodeRequest) => post<VerifyCodeResponse>('/auth/verify-code', data),
  
  /** 用户注册 */
  register: (data: RegisterRequest) => post<LoginResponse>('/auth/register', data),
  
  /** 用户登录 */
  login: (data: LoginRequest) => post<LoginResponse>('/auth/login', data),
  
  /** 刷新令牌 */
  refresh: (refreshToken: string) => post<{ accessToken: string; refreshToken: string }>('/auth/refresh', { refreshToken }),
  
  /** 用户登出 */
  logout: (refreshToken?: string) => post('/auth/logout', { refreshToken }),
  
  /** 修改密码 */
  changePassword: (oldPassword: string, newPassword: string) => post('/auth/change-password', { oldPassword, newPassword }),
  
  /** 发送密码重置验证码（忘记密码） */
  forgotPassword: (email: string) => post<{ message: string; token: string }>('/auth/forgot-password', { email }),
  
  /** 重置密码 */
  resetPassword: (token: string, code: string, newPassword: string) => post<{ message: string }>('/auth/reset-password', { token, code, newPassword }),
};

/**
 * User API
 */
export const userApi = {
  /** 获取当前用户信息 */
  getMe: () => get<UserInfo>('/users/me'),
  
  /** 更新用户资料 */
  updateProfile: (id: string, data: Partial<UserInfo>) => put<UserInfo>(`/users/${id}`, data),

  /** 上传并更新头像 */
  uploadAvatarFile: (id: string, filePath: string) =>
    uploadFile<{ id: string; avatar?: string }>(`/users/${id}/avatar/upload`, filePath),
};

// ============================================
// JWXT API - 教务系统相关
// ============================================

export interface CourseItem {
  name: string;
  teacher: string;
  classroom: string;
  weekday: number;
  startSection: number;
  endSection: number;
  weeks: string;
}

export interface CourseData {
  success: boolean;
  data?: {
    courses: CourseItem[];
    semester?: string;
  };
  error?: string;
}

export interface GradeItem {
  id: number;
  [key: string]: string | number | undefined; // 动态字段，如 "课程名称", "学分", "成绩" 等
}

export interface GradeStatistics {
  total_courses: number;
  average_score: number | null;
  weighted_average: number | null;
  total_credits: number;
  grade_distribution: Record<string, number>;
}

export interface GradeData {
  success: boolean;
  data?: {
    success: boolean;
    semester_id?: string;
    grades: GradeItem[];
    statistics: GradeStatistics;
    total_courses: number;
    message?: string;
  };
  error?: string;
}

export interface ExamItem {
  id: number;
  [key: string]: string | number | undefined; // 动态字段
  // 标准化字段
  course_name?: string;
  exam_date?: string;
  exam_time?: string;
  start_time?: string;
  end_time?: string;
  location?: string;
  seat?: string;
  exam_type?: string;
  status?: string;
}

export interface ExamData {
  success: boolean;
  data?: {
    success: boolean;
    exams: ExamItem[];
    total: number;
    semester_id?: string;
  };
  error?: string;
}

export interface SemesterInfo {
  id: string;
  name: string;
  current?: boolean;
}

export interface SemesterData {
  success: boolean;
  data?: {
    semesters: SemesterInfo[];
    current_semester?: string;
  };
  error?: string;
}

// 评教相关接口
export interface EvaluationItem {
  lesson_id: string;
  teacher_name: string;
  course_code: string;
  course_name: string;
  course_type: string;
}

export interface EvaluationPendingData {
  success: boolean;
  data?: {
    success: boolean;
    total: number;
    evaluations: EvaluationItem[];
  };
  error?: string;
}

export interface EvaluationAutoData {
  success: boolean;
  data?: {
    success: boolean;
    total: number;
    succeeded: number;
    failed: number;
    message: string;
    details?: Array<{
      lesson_id: string;
      course: string;
      teacher: string;
      success: boolean;
      message: string;
    }>;
  };
  error?: string;
}

/**
 * JWXT API - 教务系统
 */
export const jwxtApi = {
  /** 获取课程表（使用缓存） */
  getCourses: (semesterId?: string) => get<CourseData>('/jwxt/course', semesterId ? { semester_id: semesterId } : undefined),
  
  /** 刷新课程表（清除缓存重新获取） */
  refreshCourses: (semesterId?: string) => get<CourseData>('/jwxt/course/refresh', semesterId ? { semester_id: semesterId } : undefined),
  
  /** 获取成绩 */
  getGrades: (semesterId?: string) => get<GradeData>('/jwxt/grade', semesterId ? { semester_id: semesterId } : undefined),
  
  /** 获取考试安排 */
  getExams: (semesterId?: string) => get<ExamData>('/jwxt/exam', semesterId ? { semester_id: semesterId } : undefined),
  
  /** 获取学期列表 */
  getSemesters: () => get<SemesterData>('/jwxt/semester'),
  
  /** 获取教务系统用户信息 */
  getUserInfo: () => get<{ success: boolean; data?: { name: string; student_id: string; department: string; major: string; class_name: string }; error?: string }>('/jwxt/user'),
  
  /** 获取待评教列表 */
  getEvaluationPending: () => get<EvaluationPendingData>('/jwxt/evaluation/pending'),
  
  /** 一键评教 */
  autoEvaluate: () => post<EvaluationAutoData>('/jwxt/evaluation/auto'),
  
  /** 绑定教务系统账号 */
  bindAccount: (username: string, password: string) => post<{ message: string }>('/jwxt/bind', { username, password }),
  
  /** 解绑教务系统账号 */
  unbindAccount: () => post<{ message: string }>('/jwxt/unbind'),
};

// ============================================
// 公告 API
// ============================================

export interface AnnouncementItem {
  id: string;
  title: string;
  content: string;
  summary?: string;
  type: string;
  status: string;
  isPinned: boolean;
  isPopup: boolean;
  publishedAt?: string;
  expiresAt?: string;
  createdAt: string;
  author?: {
    id: string;
    username: string;
    realName?: string;
    avatar?: string;
  };
  _count?: {
    views: number;
  };
}

export interface AnnouncementListData {
  items: AnnouncementItem[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const normalizeAnnouncementItem = (raw: Record<string, unknown>): AnnouncementItem => {
  return {
    id: (raw.id ?? raw.ID ?? '') as string,
    title: (raw.title ?? raw.Title ?? '') as string,
    content: (raw.content ?? raw.Content ?? '') as string,
    summary: (raw.summary ?? raw.Summary ?? undefined) as string | undefined,
    type: (raw.type ?? raw.Type ?? 'NORMAL') as string,
    status: (raw.status ?? raw.Status ?? 'PUBLISHED') as string,
    isPinned: Boolean(raw.isPinned ?? raw.IsPinned),
    isPopup: Boolean(raw.isPopup ?? raw.IsPopup),
    publishedAt: (raw.publishedAt ?? raw.PublishedAt ?? undefined) as string | undefined,
    expiresAt: (raw.expiresAt ?? raw.ExpiresAt ?? undefined) as string | undefined,
    createdAt: (raw.createdAt ?? raw.CreatedAt ?? '') as string,
    author: (raw.author ?? raw.Author ?? undefined) as AnnouncementItem['author'],
    _count: (raw._count ?? raw.Count ?? undefined) as AnnouncementItem['_count'],
  };
};

const normalizeAnnouncementList = (raw: unknown): AnnouncementListData => {
  const payload = (raw ?? {}) as Record<string, unknown>;
  const rawItems = (payload.items ?? payload.data ?? []) as unknown[];
  const items = Array.isArray(rawItems)
    ? rawItems.map((item) => normalizeAnnouncementItem((item ?? {}) as Record<string, unknown>))
    : [];

  const total = Number(payload.total ?? items.length) || 0;
  const page = Number(payload.page ?? 1) || 1;
  const pageSize = Number((payload.pageSize ?? payload.size ?? items.length ?? 10)) || 10;
  const totalPages = Number((payload.totalPages ?? Math.ceil(total / pageSize) ?? 1)) || 1;

  return {
    items,
    total,
    page,
    pageSize,
    totalPages,
  };
};

export const announcementApi = {
  /** 获取公告列表（公开接口） */
  getList: async (params?: { page?: number; pageSize?: number; type?: string; status?: string }) => {
    const res = await get<unknown>('/announcements', {
      status: 'PUBLISHED',
      ...(params || {}),
    });
    return normalizeAnnouncementList(res);
  },

  /** 获取公告详情 */
  getById: (id: string) => get<AnnouncementItem>(`/announcements/${id}`),

  /** 获取未读公告数量 */
  getUnviewedCount: () => get<{ count: number }>('/announcements/unviewed-count'),

  /** 标记公告为已读 */
  markViewed: (id: string) => post(`/announcements/${id}/mark-viewed`),
};
