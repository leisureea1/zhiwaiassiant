// 用户类型定义
export interface User {
  id: string;
  name: string;
  studentId: string;
  avatar: string;
  college?: string;
  major?: string;
  grade?: string;
  class?: string;
}

// 用户信息存储键
export const USER_INFO_KEY = 'user_info';

// 页面路径
export const PagePath = {
  HOME: '/pages/home/index',
  LOGIN: '/pages/login/index',
  REGISTER: '/pages/register/index',
  APPS: '/pages/apps/index',
  PROFILE: '/pages/profile/index',
  GRADES: '/pages/grades/index',
  EXAMS: '/pages/exams/index',
  BUS: '/pages/bus/index',
  EVALUATION: '/pages/evaluation/index',
  AI_CHAT: '/pages/ai-chat/index',
} as const;

// 聊天消息类型
export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: number;
}

// 课程类型
export interface Course {
  id: string;
  name: string;
  teacher: string;
  location: string;
  time: string;
  dayOfWeek: number;
  startPeriod: number;
  endPeriod: number;
  weeks: string;
  color?: string;
}

// 成绩类型
export interface Grade {
  id: string;
  courseName: string;
  courseCode: string;
  credit: number;
  score: number | string;
  gradePoint: number;
  semester: string;
  type?: string;
}

// 考试类型
export interface Exam {
  id: string;
  courseName: string;
  time: string;
  location: string;
  seatNumber?: string;
  type?: string;
}

// 校车时刻
export interface BusSchedule {
  id: string;
  route: string;
  departure: string;
  arrival: string;
  departureTime: string;
  status: 'waiting' | 'departed' | 'arriving';
}
