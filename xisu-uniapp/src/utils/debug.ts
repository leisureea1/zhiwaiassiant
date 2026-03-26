/**
 * 统一的调试日志工具
 * 仅在开发环境输出日志，生产环境自动静默
 */

const IS_DEV = import.meta.env.DEV;

export const debugLog = (...args: unknown[]) => {
  if (IS_DEV) {
    // eslint-disable-next-line no-console
    console.log(...args);
  }
};

export const debugError = (...args: unknown[]) => {
  if (IS_DEV) {
    // eslint-disable-next-line no-console
    console.error(...args);
  }
};

export const debugWarn = (...args: unknown[]) => {
  if (IS_DEV) {
    // eslint-disable-next-line no-console
    console.warn(...args);
  }
};
