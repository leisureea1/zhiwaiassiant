<template>
	<view class="page">
		<!-- 头部 -->
		<view class="header">
			<view class="back-btn" @tap="handleBack">
				<text class="iconfont icon-navigate_before"></text>
			</view>
			<text class="page-title">关于我们</text>
			<view class="header-right"></view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- Logo 区域 -->
			<view class="logo-section">
				<view class="logo-wrapper">
					<image class="logo" src="/static/icons/logo.png" mode="aspectFit" />
				</view>
				<text class="app-name">知外助手</text>
				<text class="app-slogan">西外智慧服务</text>
				<view class="version-badge">
					<text class="version-text">v{{ appVersion }}</text>
				</view>
			</view>

			<!-- 应用介绍 -->
			<view class="section-card">
				<view class="card-header">
					<view class="card-icon bg-blue">
						<text class="iconfont icon-school"></text>
					</view>
					<text class="card-title">应用介绍</text>
				</view>
				<view class="card-body">
					<text class="desc-text">知外助手是专为西外打造的智慧校园工具，致力于提供便捷、高效的校园信息服务。</text>
				</view>
			</view>

			<!-- 功能特色 -->
			<view class="section-card">
				<view class="card-header">
					<view class="card-icon bg-purple">
						<text class="iconfont icon-flight_class"></text>
					</view>
					<text class="card-title">核心功能</text>
				</view>
				<view class="feature-grid">
					<view class="feature-item" v-for="(feature, index) in features" :key="index">
						<view class="feature-dot" :class="feature.color"></view>
						<text class="feature-text">{{ feature.text }}</text>
					</view>
				</view>
			</view>

			<!-- 联系我们 -->
			<view class="section-card">
				<view class="card-header">
					<view class="card-icon bg-teal">
						<text class="iconfont icon-campaign"></text>
					</view>
					<text class="card-title">联系与反馈</text>
				</view>
				<view class="contact-list">
					<view class="contact-item" @tap="copyEmail">
						<text class="contact-label">邮箱</text>
						<text class="contact-value">leisureea@gmail.com</text>
						<text class="iconfont icon-event_note copy-icon"></text>
					</view>
					<view class="contact-divider"></view>
					<view class="contact-item">
						<text class="contact-label">开发</text>
						<text class="contact-value">Leisure</text>
					</view>
				</view>
			</view>

			<!-- 法律信息 -->
			<view class="section-card">
				<view class="card-header">
					<view class="card-icon bg-gray">
						<text class="iconfont icon-assignment_ind"></text>
					</view>
					<text class="card-title">法律信息</text>
				</view>
				<view class="legal-list">
					<view class="legal-item" @tap="navigateTo('/pages/privacy-policy/index')">
						<text class="legal-label">隐私政策</text>
						<text class="arrow">›</text>
					</view>
					<view class="contact-divider"></view>
					<view class="legal-item" @tap="navigateTo('/pages/service-agreement/index')">
						<text class="legal-label">用户服务协议</text>
						<text class="arrow">›</text>
					</view>
				</view>
			</view>

			<!-- 底部 -->
			<view class="footer">
				<text class="footer-text">© 2026 Leisure</text>
				<text class="footer-text">Made with ❤️ in Xi'an</text>
			</view>
		</scroll-view>
	</view>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';

const appVersion = ref('2.4.1');

const features = ref([
	{ text: '课程表查询', color: 'dot-blue' },
	{ text: '成绩查询', color: 'dot-green' },
	{ text: '考试安排', color: 'dot-orange' },
	{ text: '一键评教', color: 'dot-purple' },
	{ text: '校车时刻', color: 'dot-teal' },
	{ text: 'AI 智能助手', color: 'dot-indigo' },
]);

const handleBack = () => {
	safeNavigateBack();
};

const copyEmail = () => {
	uni.setClipboardData({
		data: 'zhiwai@xisu.edu.cn',
		success: () => {
			uni.showToast({ title: '邮箱已复制', icon: 'success' });
		}
	});
};

const navigateTo = (url: string) => {
	uni.navigateTo({ url });
};
</script>

<style lang="scss" scoped>
.page {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	background-color: $bg-light;
}

.header {
	position: sticky;
	top: 0;
	z-index: 10;
	display: flex;
	align-items: center;
	padding: 24rpx 32rpx;
	padding-top: calc(var(--status-bar-height) + 48rpx);
	background: rgba(248, 250, 252, 0.9);
	backdrop-filter: blur(20rpx);
}

.back-btn {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	background-color: #fff;
	box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.05);
	display: flex;
	align-items: center;
	justify-content: center;

	.iconfont {
		font-size: 40rpx;
		color: $text-primary;
	}
}

.page-title {
	flex: 1;
	text-align: center;
	font-size: 36rpx;
	font-weight: 700;
	color: $text-primary;
}

.header-right {
	width: 72rpx;
}

.content {
	flex: 1;
	padding: 0 32rpx;
}

/* Logo 区域 */
.logo-section {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 48rpx 0 40rpx;
}

.logo-wrapper {
	width: 160rpx;
	height: 160rpx;
	border-radius: 48rpx;
	background-color: #fff;
	box-shadow: 0 8rpx 32rpx rgba(59, 130, 246, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 24rpx;
	overflow: hidden;
}

.logo {
	width: 120rpx;
	height: 120rpx;
}

.app-name {
	font-size: 44rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 8rpx;
}

.app-slogan {
	font-size: 26rpx;
	color: $text-secondary;
	margin-bottom: 20rpx;
}

.version-badge {
	padding: 8rpx 28rpx;
	background-color: rgba(59, 130, 246, 0.1);
	border-radius: 32rpx;
}

.version-text {
	font-size: 24rpx;
	font-weight: 500;
	color: $primary;
}

/* 通用卡片 */
.section-card {
	background-color: #fff;
	border-radius: 48rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
	border: 2rpx solid rgba(0, 0, 0, 0.04);
	overflow: hidden;
}

.card-header {
	display: flex;
	align-items: center;
	padding: 32rpx 40rpx 0;
}

.card-icon {
	width: 64rpx;
	height: 64rpx;
	border-radius: 20rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-right: 20rpx;

	.iconfont {
		font-size: 32rpx;
	}

	&.bg-blue {
		background-color: rgba(59, 130, 246, 0.1);
		.iconfont { color: #3b82f6; }
	}
	&.bg-purple {
		background-color: rgba(139, 92, 246, 0.1);
		.iconfont { color: #8b5cf6; }
	}
	&.bg-teal {
		background-color: rgba(20, 184, 166, 0.1);
		.iconfont { color: #14b8a6; }
	}
	&.bg-gray {
		background-color: rgba(107, 114, 128, 0.1);
		.iconfont { color: #6b7280; }
	}
}

.card-title {
	font-size: 32rpx;
	font-weight: 600;
	color: $text-primary;
}

.card-body {
	padding: 24rpx 40rpx 32rpx;
}

.desc-text {
	font-size: 28rpx;
	color: $text-secondary;
	line-height: 1.8;
}

/* 功能网格 */
.feature-grid {
	display: flex;
	flex-wrap: wrap;
	padding: 24rpx 40rpx 32rpx;
	gap: 20rpx;
}

.feature-item {
	display: flex;
	align-items: center;
	width: calc(50% - 10rpx);
	padding: 16rpx 0;
}

.feature-dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	margin-right: 16rpx;
	flex-shrink: 0;

	&.dot-blue { background-color: #3b82f6; }
	&.dot-green { background-color: #10b981; }
	&.dot-orange { background-color: #f97316; }
	&.dot-purple { background-color: #8b5cf6; }
	&.dot-teal { background-color: #14b8a6; }
	&.dot-indigo { background-color: #6366f1; }
}

.feature-text {
	font-size: 28rpx;
	color: $text-primary;
	font-weight: 500;
}

/* 联系信息 */
.contact-list {
	padding: 16rpx 40rpx 24rpx;
}

.contact-item {
	display: flex;
	align-items: center;
	padding: 20rpx 0;

	&:active {
		opacity: 0.7;
	}
}

.contact-label {
	font-size: 28rpx;
	color: $text-secondary;
	width: 140rpx;
}

.contact-value {
	flex: 1;
	font-size: 28rpx;
	color: $text-primary;
	font-weight: 500;
}

.copy-icon {
	font-size: 32rpx;
	color: $text-light;
}

.contact-divider {
	height: 2rpx;
	background-color: $border-color;
	margin-left: 140rpx;
}

/* 法律信息 */
.legal-list {
	padding: 8rpx 40rpx 16rpx;
}

.legal-item {
	display: flex;
	align-items: center;
	padding: 24rpx 0;

	&:active {
		opacity: 0.7;
	}
}

.legal-label {
	flex: 1;
	font-size: 28rpx;
	color: $text-primary;
	font-weight: 500;
}

.arrow {
	font-size: 40rpx;
	color: $text-light;
}

/* 底部 */
.footer {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 48rpx 0 80rpx;
	gap: 8rpx;
}

.footer-text {
	font-size: 24rpx;
	color: $text-light;
}
</style>
