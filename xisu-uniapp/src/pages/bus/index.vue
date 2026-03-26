<template>
	<view class="bus-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-content">
				<view class="header-left">
					<view class="back-link" @tap="goBack">
						<text class="iconfont icon-arrow_back"></text>
						<text class="back-text">返回</text>
					</view>
					<text class="page-title">校区通勤班车</text>
					<text class="page-subtitle">智外助手 · 直达通勤班次</text>
				</view>
				<image class="avatar" :src="user.avatar" mode="aspectFill" />
			</view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- 方向选择 -->
			<view class="direction-tabs">
				<view 
					class="tab-item" 
					:class="{ active: direction === 'A' }"
					@tap="direction = 'A'"
				>
					<text class="iconfont icon-map"></text>
					<text class="tab-text">前往A校区</text>
				</view>
				<view 
					class="tab-item" 
					:class="{ active: direction === 'B' }"
					@tap="direction = 'B'"
				>
					<text class="iconfont icon-map"></text>
					<text class="tab-text">前往B校区</text>
				</view>
			</view>

			<!-- 最近班次卡片 -->
			<view class="hero-card">
				<view class="hero-bg-circle top"></view>
				<view class="hero-bg-circle bottom"></view>
				
				<view class="hero-content">
					<view class="hero-header">
						<view class="hero-left">
							<view class="next-badge">
								<view class="pulse-dot"></view>
								<text class="badge-text">最近班次 (Next)</text>
							</view>
							<view class="next-time">
								<text class="time-value">14:30</text>
								<text class="time-label">发车</text>
							</view>
						</view>
						<view class="countdown-box">
							<view class="countdown-value">
								<text class="count-num">12</text>
								<text class="count-unit">min</text>
							</view>
							<text class="countdown-label">倒计时</text>
						</view>
					</view>
					
					<view class="hero-info">
						<view class="bus-icon-wrap">
							<text class="iconfont icon-directions_bus"></text>
						</view>
						<view class="bus-info">
							<text class="bus-name">校区专线 (A ⇄ B)</text>
							<text class="bus-eta">预计 15:15 到达 · 耗时 45min</text>
						</view>
						<view class="seat-info">
							<view class="seat-badge">有位</view>
							<text class="seat-count">剩余 18 座</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 班次列表 -->
			<view class="schedule-section">
				<view class="schedule-header">
					<view class="header-title">
						<text class="iconfont icon-calendar_today"></text>
						<text class="title-text">今日班次计划</text>
					</view>
					<text class="header-count">共 12 个班次</text>
				</view>

				<view class="schedule-list">
					<view 
						v-for="(bus, index) in busSchedule" 
						:key="index"
						class="schedule-item"
						:class="{ disabled: bus.status === '已满' }"
					>
						<view class="item-left">
							<text class="bus-time">{{ bus.time }}</text>
							<view class="divider"></view>
							<view class="arrive-info">
								<text class="arrive-label">预计到达</text>
								<text class="arrive-time">{{ bus.arrive }}</text>
							</view>
						</view>
						<view class="item-right">
							<view class="status-badge" :class="bus.statusColor">
								<text class="status-text">{{ bus.status }}</text>
							</view>
							<text class="arrow">›</text>
						</view>
					</view>
				</view>

				<view class="schedule-tip">
					<text class="tip-text">班次仅供参考，请提前5分钟到达上车点</text>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { getUserInfo } from '@/services/apiService';
import { recordAppUsage } from '@/utils/appUsage';

const user = ref({ avatar: '/static/default-avatar.png', name: '用户' });

onMounted(() => {
	recordAppUsage('bus');
	const userInfo = getUserInfo<{ avatar?: string; realName?: string; username?: string }>();
	if (userInfo) {
		user.value = {
			avatar: userInfo.avatar || '/static/default-avatar.png',
			name: userInfo.realName || userInfo.username || '用户'
		};
	}
});
const direction = ref('A');

const busSchedule = ref([
	{ time: '15:00', arrive: '15:45', status: '有位', statusColor: 'green' },
	{ time: '15:30', arrive: '16:15', status: '有位', statusColor: 'green' },
	{ time: '16:00', arrive: '16:45', status: '紧俏', statusColor: 'orange' },
	{ time: '16:30', arrive: '17:15', status: '已满', statusColor: 'red' },
	{ time: '17:00', arrive: '17:45', status: '有位', statusColor: 'green' }
]);

const goBack = () => {
	safeNavigateBack();
};
</script>

<style lang="scss" scoped>
.bus-page {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	background-color: $bg-light;
}

.header {
	position: sticky;
	top: 0;
	z-index: 50;
	background: rgba(248, 250, 252, 0.8);
	backdrop-filter: blur(20rpx);
	border-bottom: 2rpx solid $border-color;
	padding: 24rpx 32rpx;
	padding-top: calc(var(--status-bar-height) + 48rpx);
}

.header-content {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
}

.header-left {
	display: flex;
	flex-direction: column;
}

.back-link {
	display: flex;
	align-items: center;
	margin-bottom: 8rpx;
	
	.icon {
		font-size: 28rpx;
		margin-right: 8rpx;
		color: $text-secondary;
	}
	
	.back-text {
		font-size: 28rpx;
		color: $text-secondary;
	}
	
	&:active {
		opacity: 0.7;
	}
}

.page-title {
	font-size: 40rpx;
	font-weight: 700;
	color: $text-primary;
}

.page-subtitle {
	font-size: 28rpx;
	color: $text-secondary;
	margin-top: 8rpx;
}

.avatar {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	border: 2rpx solid $border-color;
}

.content {
	flex: 1;
	padding: 32rpx;
}

.direction-tabs {
	display: flex;
	background-color: #fff;
	border-radius: 32rpx;
	padding: 8rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
	margin-bottom: 24rpx;
}

.tab-item {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 8rpx;
	padding: 24rpx;
	border-radius: 24rpx;
	
	.icon {
		font-size: 24rpx;
	}
	
	.tab-text {
		font-size: 28rpx;
		font-weight: 500;
		color: $text-secondary;
	}
	
	&.active {
		background-color: rgba(59, 130, 246, 0.1);
		
		.tab-text {
			color: $primary;
			font-weight: 700;
		}
	}
}

.hero-card {
	position: relative;
	background: linear-gradient(135deg, #2563EB 0%, #3B82F6 100%);
	border-radius: 48rpx;
	padding: 48rpx;
	margin-bottom: 32rpx;
	overflow: hidden;
	box-shadow: 0 16rpx 48rpx rgba(59, 130, 246, 0.2);
}

.hero-bg-circle {
	position: absolute;
	border-radius: 50%;
	background-color: rgba(255,255,255,0.1);
	filter: blur(40rpx);
	
	&.top {
		width: 384rpx;
		height: 384rpx;
		top: -160rpx;
		right: -80rpx;
	}
	
	&.bottom {
		width: 256rpx;
		height: 256rpx;
		bottom: -128rpx;
		left: -64rpx;
	}
}

.hero-content {
	position: relative;
	z-index: 10;
}

.hero-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	margin-bottom: 48rpx;
}

.hero-left {
	display: flex;
	flex-direction: column;
}

.next-badge {
	display: inline-flex;
	align-items: center;
	gap: 12rpx;
	padding: 12rpx 24rpx;
	background-color: rgba(255,255,255,0.2);
	backdrop-filter: blur(10rpx);
	border-radius: 32rpx;
	border: 2rpx solid rgba(255,255,255,0.1);
	margin-bottom: 24rpx;
}

.pulse-dot {
	width: 12rpx;
	height: 12rpx;
	background-color: #86efac;
	border-radius: 50%;
	animation: pulse 1.5s infinite;
}

@keyframes pulse {
	0%, 100% { opacity: 1; }
	50% { opacity: 0.5; }
}

.badge-text {
	font-size: 24rpx;
	color: #fff;
	font-weight: 500;
}

.next-time {
	display: flex;
	align-items: baseline;
	gap: 8rpx;
}

.time-value {
	font-size: 80rpx;
	font-weight: 700;
	color: #fff;
}

.time-label {
	font-size: 32rpx;
	color: rgba(255,255,255,0.8);
}

.countdown-box {
	background-color: rgba(255,255,255,0.1);
	backdrop-filter: blur(10rpx);
	border: 2rpx solid rgba(255,255,255,0.1);
	border-radius: 32rpx;
	padding: 24rpx;
	text-align: center;
}

.countdown-value {
	display: flex;
	align-items: baseline;
	justify-content: center;
}

.count-num {
	font-size: 56rpx;
	font-weight: 700;
	color: #fff;
}

.count-unit {
	font-size: 24rpx;
	color: rgba(255,255,255,0.8);
	margin-left: 4rpx;
}

.countdown-label {
	font-size: 24rpx;
	color: rgba(255,255,255,0.8);
	margin-top: 8rpx;
}

.hero-info {
	display: flex;
	align-items: center;
	background-color: rgba(255,255,255,0.1);
	backdrop-filter: blur(10rpx);
	border: 2rpx solid rgba(255,255,255,0.1);
	border-radius: 24rpx;
	padding: 24rpx;
}

.bus-icon-wrap {
	width: 80rpx;
	height: 80rpx;
	background-color: #fff;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-right: 24rpx;
	box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.1);
	
	.icon {
		font-size: 40rpx;
	}
}

.bus-info {
	flex: 1;
	display: flex;
	flex-direction: column;
}

.bus-name {
	font-size: 28rpx;
	font-weight: 700;
	color: #fff;
}

.bus-eta {
	font-size: 24rpx;
	color: rgba(255,255,255,0.9);
	margin-top: 4rpx;
}

.seat-info {
	display: flex;
	flex-direction: column;
	align-items: flex-end;
}

.seat-badge {
	padding: 8rpx 16rpx;
	background-color: #4ade80;
	color: #1e3a8a;
	font-size: 20rpx;
	font-weight: 700;
	border-radius: 8rpx;
	text-transform: uppercase;
}

.seat-count {
	font-size: 20rpx;
	color: rgba(255,255,255,0.7);
	margin-top: 8rpx;
}

.schedule-section {
	background-color: #fff;
	border-radius: 48rpx;
	padding: 40rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
}

.schedule-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 32rpx;
}

.header-title {
	display: flex;
	align-items: center;
	gap: 16rpx;
	
	.icon {
		font-size: 36rpx;
		color: $primary;
	}
	
	.title-text {
		font-size: 36rpx;
		font-weight: 700;
		color: $text-primary;
	}
}

.header-count {
	font-size: 24rpx;
	color: $text-secondary;
}

.schedule-list {
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.schedule-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
	background-color: #fff;
	border: 2rpx solid $border-color;
	border-radius: 32rpx;
	padding: 32rpx;
	box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.02);
	
	&.disabled {
		opacity: 0.6;
	}
	
	&:active:not(.disabled) {
		transform: scale(0.98);
	}
}

.item-left {
	display: flex;
	align-items: center;
	gap: 24rpx;
}

.bus-time {
	font-size: 40rpx;
	font-weight: 700;
	color: $text-primary;
}

.divider {
	width: 2rpx;
	height: 64rpx;
	background-color: $border-color;
}

.arrive-info {
	display: flex;
	flex-direction: column;
}

.arrive-label {
	font-size: 24rpx;
	color: $text-light;
}

.arrive-time {
	font-size: 28rpx;
	font-weight: 500;
	color: $text-primary;
}

.item-right {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.status-badge {
	padding: 8rpx 24rpx;
	border-radius: 32rpx;
	
	&.green {
		background-color: rgba(16, 185, 129, 0.05);
		
		.status-text { color: $success; }
	}
	
	&.orange {
		background-color: rgba(249, 115, 22, 0.05);
		
		.status-text { color: #f97316; }
	}
	
	&.red {
		background-color: rgba(239, 68, 68, 0.05);
		
		.status-text { color: $danger; }
	}
}

.status-text {
	font-size: 24rpx;
	font-weight: 500;
}

.arrow {
	font-size: 36rpx;
	color: $text-light;
}

.schedule-tip {
	text-align: center;
	padding-top: 32rpx;
}

.tip-text {
	font-size: 20rpx;
	color: $text-light;
}
</style>
