<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app';
import { debugLog } from '@/utils/debug';

onLaunch(() => {
	debugLog('知外助手启动');
});

onShow(() => {
	debugLog('App Show');
});

onHide(() => {
	debugLog('App Hide');
});
</script>

<style lang="scss">
	/* 引入 iconfont */
	@import '@/static/iconfont/iconfont.css';
	
	/* 全局公共样式 */
	page {
		background-color: #F8FAFC;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
		color: #1E293B;
		box-sizing: border-box;
	}
	
	view, text, image, scroll-view {
		box-sizing: border-box;
	}
	
	/* 安全区域 */
	.safe-area-bottom {
		padding-bottom: env(safe-area-inset-bottom);
	}
	
	/* 滚动条隐藏 */
	::-webkit-scrollbar {
		display: none;
		width: 0;
		height: 0;
		background: transparent;
	}
	
	/* 禁用按钮默认样式 */
	button {
		margin: 0;
		padding: 0;
		background: none;
		border: none;
		line-height: inherit;
		font-size: inherit;
		
		&::after {
			display: none;
		}
	}
	
	/* 图片默认样式 */
	image {
		display: block;
	}
	
	/* 通用动画 */
	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}
	
	@keyframes slideUp {
		from { 
			opacity: 0;
			transform: translateY(20rpx);
		}
		to { 
			opacity: 1;
			transform: translateY(0);
		}
	}
	
	.animate-fade-in {
		animation: fadeIn 0.3s ease-out;
	}
	
	.animate-slide-up {
		animation: slideUp 0.3s ease-out;
	}
</style>
