/**
 * 密码校验与强度计算工具函数
 * 统一各页面的密码验证逻辑
 */

/** 密码强度等级：0=无, 1=弱, 2=中, 3=强, 4=很强 */
export type PasswordStrength = 0 | 1 | 2 | 3 | 4;

/**
 * 计算密码强度等级 (0-4)
 * 规则：
 *   - 长度 >= 8  → +1
 *   - 长度 >= 12 → +1
 *   - 包含小写字母 → +1
 *   - 包含大写字母 → +1
 *   - 包含数字 → +1
 *   - 包含特殊字符 → +1
 * 映射到 1-4 级：
 *   <=2 → 1(弱), <=3 → 2(中), <=4 → 3(强), 5-6 → 4(很强)
 */
export function calcPasswordStrength(password: string): PasswordStrength {
  if (!password || password.length < 8) return 0;

  let score = 0;
  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[a-z]/.test(password)) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[0-9]/.test(password)) score++;
  if (/[^A-Za-z0-9]/.test(password)) score++;

  if (score <= 2) return 1;
  if (score <= 3) return 2;
  if (score <= 4) return 3;
  return 4;
}

/** 密码强度中文描述 */
export const strengthLabels: Record<PasswordStrength, string> = {
  0: '',
  1: '弱',
  2: '中',
  3: '强',
  4: '很强',
};

/** 密码强度 CSS 类名 */
export const strengthClasses: Record<PasswordStrength, string> = {
  0: '',
  1: 'weak',
  2: 'medium',
  3: 'strong',
  4: 'very-strong',
};

/** 密码强度颜色 */
export const strengthColors: Record<PasswordStrength, string> = {
  0: '',
  1: '#ef4444',
  2: '#f59e0b',
  3: '#10b981',
  4: '#3b82f6',
};

/**
 * 校验密码格式（至少8位，包含大小写字母和数字）
 * @returns 错误信息，空字符串表示通过
 */
export function validatePasswordFormat(password: string): string {
  if (!password) return '请输入密码';
  if (password.length < 8) return '密码长度至少为8位';
  if (!/[a-z]/.test(password)) return '密码需包含小写字母';
  if (!/[A-Z]/.test(password)) return '密码需包含大写字母';
  if (!/[0-9]/.test(password)) return '密码需包含数字';
  return '';
}

/**
 * 校验两次密码是否一致
 * @returns 错误信息，空字符串表示通过
 */
export function validatePasswordMatch(password: string, confirmPassword: string): string {
  if (!confirmPassword) return '';
  if (password !== confirmPassword) return '两次输入的密码不一致';
  return '';
}
