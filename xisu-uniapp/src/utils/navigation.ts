/**
 * 安全的导航返回方法
 * 如果当前页面是页面栈中的第一个页面（无法返回），则跳转到首页
 * @param fallbackUrl 回退页面路径，默认为首页（switchTab）
 */
export const safeNavigateBack = (fallbackUrl?: string) => {
  const pages = getCurrentPages();
  if (pages.length > 1) {
    uni.navigateBack();
  } else {
    // 无法返回，跳转到指定页面或首页
    if (fallbackUrl) {
      uni.redirectTo({ url: fallbackUrl });
    } else {
      uni.switchTab({ url: '/pages/home/index' });
    }
  }
};
