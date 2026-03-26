/**
 * 应用使用记录工具
 */

const STORAGE_KEY = 'recent_app_usage';

/**
 * 记录应用使用
 * @param appId 应用ID
 */
export const recordAppUsage = (appId: string): void => {
	try {
		const recentUsageStr = uni.getStorageSync(STORAGE_KEY);
		const recentUsage = recentUsageStr ? JSON.parse(recentUsageStr) : {};
		recentUsage[appId] = Date.now();
		uni.setStorageSync(STORAGE_KEY, JSON.stringify(recentUsage));
	} catch (error) {
		console.error('[AppUsage] Failed to record app usage:', error);
	}
};

/**
 * 获取应用使用记录
 * @returns 应用使用记录对象，key为应用ID，value为最后使用时间戳
 */
export const getAppUsage = (): Record<string, number> => {
	try {
		const recentUsageStr = uni.getStorageSync(STORAGE_KEY);
		return recentUsageStr ? JSON.parse(recentUsageStr) : {};
	} catch (error) {
		console.error('[AppUsage] Failed to get app usage:', error);
		return {};
	}
};

/**
 * 清除应用使用记录
 */
export const clearAppUsage = (): void => {
	try {
		uni.removeStorageSync(STORAGE_KEY);
	} catch (error) {
		console.error('[AppUsage] Failed to clear app usage:', error);
	}
};
