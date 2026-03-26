<template>
	<view class="subscription-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-content">
				<view class="header-left">
					<view class="back-btn" @tap="goBack">
						<text class="iconfont icon-arrow_back"></text>
					</view>
				</view>
				<text class="page-title">成绩订阅</text>
				<view style="width: 80rpx;"></view>
			</view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- 说明卡片 -->
			<view class="info-card">
				<view class="info-icon">
					<text class="iconfont icon-event_note"></text>
				</view>
				<text class="info-title">开启成绩订阅后</text>
				<text class="info-desc">系统将每小时自动检查您的本学期成绩，如果发现成绩变化，将通过您的注册邮箱发送通知</text>
			</view>

			<!-- 开关卡片 -->
			<view class="card">
				<view class="card-header">
					<view class="card-icon-wrap">
						<text class="iconfont icon-notifications_active"></text>
					</view>
					<view class="card-info">
						<text class="card-title">成绩变化通知</text>
						<text class="card-desc">有新成绩时通过邮箱提醒</text>
					</view>
					<switch 
						:checked="subscription.enabled" 
						color="#3B82F6"
						@change="toggleSubscription"
						:disabled="loading || !canSubscribe"
					/>
				</view>
				<view v-if="subscription.enabled" class="status-bar active">
					<view class="status-dot"></view>
					<text class="status-text">订阅已开启，正在监控成绩变化</text>
				</view>
				<view v-else class="status-bar inactive">
					<view class="status-dot inactive-dot"></view>
					<text class="status-text">订阅未开启</text>
				</view>
			</view>

			<!-- 预检查警告 -->
			<view v-if="warningMsg" class="warning-card">
				<text class="iconfont icon-warning"></text>
				<text class="warning-text">{{ warningMsg }}</text>
				<view v-if="!isJwxtBound" class="warning-action" @tap="goBindJwxt">
					<text class="warning-link">去绑定</text>
				</view>
			</view>

			<!-- 订阅详情 -->
			<view v-if="subscription.enabled && (subscription.totalNotified ?? 0) > 0" class="card">
				<text class="detail-title">订阅详情</text>
				<view class="detail-row">
					<text class="detail-label">累计通知次数</text>
					<text class="detail-value">{{ subscription.totalNotified }} 次</text>
				</view>
				<view v-if="subscription.lastNotifiedAt" class="detail-row">
					<text class="detail-label">上次通知时间</text>
					<text class="detail-value">{{ formatTime(subscription.lastNotifiedAt) }}</text>
				</view>
				<view v-if="subscription.lastCheckedAt" class="detail-row">
					<text class="detail-label">上次检查时间</text>
					<text class="detail-value">{{ formatTime(subscription.lastCheckedAt) }}</text>
				</view>
				<view v-if="subscription.semesterId" class="detail-row">
					<text class="detail-label">监控学期</text>
					<text class="detail-value">本学期</text>
				</view>
			</view>

			<!-- 功能说明 -->
			<view class="card">
				<text class="detail-title">使用说明</text>
				<view class="faq-item">
					<text class="faq-q">Q: 检查频率是多少？</text>
					<text class="faq-a">系统每小时自动检查一次成绩变化。</text>
				</view>
				<view class="faq-item">
					<text class="faq-q">Q: 成绩通知发送到哪里？</text>
					<text class="faq-a">通知将发送到您注册时使用的邮箱地址。</text>
				</view>
				<view class="faq-item">
					<text class="faq-q">Q: 如何关闭通知？</text>
					<text class="faq-a">关闭上方开关即可，关闭后不再检查成绩。</text>
				</view>
				<view class="faq-item">
					<text class="faq-q">Q: 需要什么前提条件？</text>
					<text class="faq-a">需要先绑定教务系统账号，并确保注册邮箱可用。</text>
				</view>
			</view>

			<view class="bottom-spacer"></view>
		</scroll-view>
	</view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { gradeSubscriptionApi } from '@/services/apiService';
import { getUserInfo } from '@/services/apiService';

const subscription = ref<{
	enabled: boolean;
	lastCheckedAt?: string;
	lastNotifiedAt?: string;
	totalNotified?: number;
	semesterId?: string;
}>({ enabled: false });

const loading = ref(false);
const isJwxtBound = ref(false);
const hasEmail = ref(false);

const warningMsg = computed(() => {
	if (!isJwxtBound.value) return '请先绑定教务系统账号才能开启成绩订阅';
	if (!hasEmail.value) return '请先设置邮箱才能接收成绩通知';
	return '';
});

const canSubscribe = computed(() => isJwxtBound.value && hasEmail.value);

const loadSubscriptionStatus = async () => {
	try {
		const res = await gradeSubscriptionApi.getStatus();
		subscription.value = {
			enabled: res.enabled,
			lastCheckedAt: res.lastCheckedAt,
			lastNotifiedAt: res.lastNotifiedAt,
			totalNotified: res.totalNotified,
			semesterId: res.semesterId,
		};
	} catch {
		// silently fail
	}
};

const loadUserInfo = () => {
	const userInfo = getUserInfo<{
		jwxtBound?: boolean;
		email?: string;
	}>();
	if (userInfo) {
		isJwxtBound.value = !!userInfo.jwxtBound;
		hasEmail.value = !!userInfo.email;
	}
};

const toggleSubscription = async () => {
	if (!canSubscribe.value || loading.value) return;

	loading.value = true;
	try {
		const newEnabled = !subscription.value.enabled;
		const res = await gradeSubscriptionApi.update(newEnabled);
		subscription.value.enabled = res.enabled;
		uni.showToast({
			title: res.message || (newEnabled ? '已开启订阅' : '已关闭订阅'),
			icon: 'none',
		});
	} catch (err: any) {
		uni.showToast({
			title: err.message || '操作失败',
			icon: 'none',
		});
	} finally {
		loading.value = false;
	}
};

const formatTime = (timeStr?: string) => {
	if (!timeStr) return '-';
	const date = new Date(timeStr);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffMin = Math.floor(diffMs / 60000);

	if (diffMin < 1) return '刚刚';
	if (diffMin < 60) return `${diffMin} 分钟前`;

	const diffHour = Math.floor(diffMin / 60);
	if (diffHour < 24) return `${diffHour} 小时前`;

	const diffDay = Math.floor(diffHour / 24);
	if (diffDay < 7) return `${diffDay} 天前`;

	return `${date.getMonth() + 1}/${date.getDate()} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
};

const goBack = () => {
	uni.navigateBack({
		fail: () => {
			uni.switchTab({ url: '/pages/apps/index' });
		}
	});
};

const goBindJwxt = () => {
	uni.navigateTo({ url: '/pages/profile/bind-jwxt' });
};

onMounted(() => {
	loadUserInfo();
	loadSubscriptionStatus();
});
</script>

<style lang="scss" scoped>
.subscription-page {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	background-color: $bg-light;
}

.header {
	position: sticky;
	top: 0;
	z-index: 40;
	background: rgba(248, 250, 252, 0.9);
	backdrop-filter: blur(20rpx);
	border-bottom: 2rpx solid $border-color;
	padding: 24rpx 32rpx;
	padding-top: calc(var(--status-bar-height) + 48rpx);
}

.header-content {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.header-left {
	display: flex;
	align-items: center;
}

.back-btn {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;

	.icon {
		font-size: 36rpx;
		color: $text-primary;
	}

	&:active {
		background-color: rgba(0, 0, 0, 0.05);
	}
}

.page-title {
	font-size: 36rpx;
	font-weight: 700;
	color: $text-primary;
}

.content {
	flex: 1;
	padding: 24rpx 32rpx;
}

.info-card {
	background: linear-gradient(135deg, #3B82F6 0%, #6366F1 100%);
	border-radius: 32rpx;
	padding: 48rpx 40rpx;
	margin-bottom: 24rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
}

.info-icon {
	width: 96rpx;
	height: 96rpx;
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.2);
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 24rpx;

	.icon {
		font-size: 48rpx;
		color: #fff;
	}
}

.info-title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	margin-bottom: 16rpx;
}

.info-desc {
	font-size: 26rpx;
	color: rgba(255, 255, 255, 0.85);
	text-align: center;
	line-height: 1.6;
}

.card {
	background: #fff;
	border-radius: 24rpx;
	padding: 36rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
	border: 2rpx solid rgba(0, 0, 0, 0.04);
}

.card-header {
	display: flex;
	align-items: center;
	gap: 20rpx;
}

.card-icon-wrap {
	width: 80rpx;
	height: 80rpx;
	border-radius: 20rpx;
	background: #EFF6FF;
	display: flex;
	align-items: center;
	justify-content: center;

	.icon {
		font-size: 40rpx;
		color: #3B82F6;
	}
}

.card-info {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: 6rpx;
}

.card-title {
	font-size: 30rpx;
	font-weight: 600;
	color: $text-primary;
}

.card-desc {
	font-size: 24rpx;
	color: $text-secondary;
}

.status-bar {
	display: flex;
	align-items: center;
	gap: 12rpx;
	margin-top: 24rpx;
	padding-top: 24rpx;
	border-top: 2rpx solid $border-color;

	&.active {
		background: #F0FDF4;
		margin: 24rpx -36rpx 0;
		padding: 20rpx 36rpx;
		border-top: none;
		border-radius: 0 0 24rpx 24rpx;
	}
}

.status-dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	background: #22C55E;

	&.inactive-dot {
		background: #D1D5DB;
	}
}

.status-text {
	font-size: 24rpx;
	color: $text-secondary;
}

.warning-card {
	background: #FFFBEB;
	border-radius: 24rpx;
	padding: 28rpx 32rpx;
	margin-bottom: 24rpx;
	display: flex;
	align-items: center;
	gap: 16rpx;
	border: 2rpx solid #FDE68A;

	.icon {
		font-size: 36rpx;
		color: #F59E0B;
		flex-shrink: 0;
	}

	.warning-text {
		font-size: 26rpx;
		color: #92400E;
		flex: 1;
	}
}

.warning-action {
	flex-shrink: 0;
}

.warning-link {
	font-size: 26rpx;
	color: #3B82F6;
	font-weight: 600;
}

.detail-title {
	font-size: 30rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 24rpx;
}

.detail-row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 16rpx 0;

	&:not(:last-child) {
		border-bottom: 2rpx solid $border-color;
	}
}

.detail-label {
	font-size: 26rpx;
	color: $text-secondary;
}

.detail-value {
	font-size: 26rpx;
	color: $text-primary;
	font-weight: 500;
}

.faq-item {
	margin-bottom: 24rpx;

	&:last-child {
		margin-bottom: 0;
	}
}

.faq-q {
	font-size: 26rpx;
	font-weight: 600;
	color: $text-primary;
	margin-bottom: 8rpx;
	display: block;
}

.faq-a {
	font-size: 24rpx;
	color: $text-secondary;
	line-height: 1.6;
	display: block;
	padding-left: 16rpx;
	border-left: 4rpx solid #3B82F6;
}

.bottom-spacer {
	height: 120rpx;
}
</style>
