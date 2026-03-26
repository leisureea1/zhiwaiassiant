<template>
	<view class="apps-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-content">
				<view class="header-left">
					<text class="page-title">应用中心</text>
					<text class="page-subtitle">发现更多校园服务</text>
				</view>
				<view class="header-right">
					<view class="search-btn" @tap="handleSearch">
						<text class="iconfont icon-event_note"></text>
					</view>
					<image class="avatar" :src="user.avatar" mode="aspectFill" />
				</view>
			</view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- 最近使用 -->
			<view class="section">
				<view class="section-header">
					<text class="section-title">最近使用</text>
				</view>
				<view class="app-grid">
					<view 
						v-for="app in recentApps" 
						:key="app.id"
						class="app-item" 
						@tap="navigateTo(app.path, app.id)"
					>
						<view class="app-icon" :class="app.iconBg">
							<text class="iconfont" :class="app.icon"></text>
						</view>
						<text class="app-label">{{ app.label }}</text>
					</view>
				</view>
			</view>

			<!-- 教务教学 -->
			<view class="section card-section">
				<view class="section-header-inner">
					<view class="indicator bg-primary"></view>
					<text class="section-title-inner">教务教学</text>
				</view>
				<view class="mini-app-grid">
					<view class="mini-app-item" @tap="navigateTo('/pages/exams/index', 'exams')">
						<view class="mini-app-icon">
							<text class="iconfont icon-fact_check"></text>
						</view>
						<text class="mini-app-label">考试安排</text>
					</view>
					<view class="mini-app-item" @tap="navigateTo('/pages/grades/index', 'grades')">
						<view class="mini-app-icon">
							<text class="iconfont icon-school"></text>
						</view>
						<text class="mini-app-label">成绩查询</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-event_note"></text>
						</view>
						<text class="mini-app-label">选课系统</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-workspace_premium"></text>
						</view>
						<text class="mini-app-label">素拓分</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-menu_book"></text>
						</view>
						<text class="mini-app-label">教材预订</text>
					</view>
					<view class="mini-app-item" @tap="navigateTo('/pages/evaluation/index', 'evaluation')">
						<view class="mini-app-icon">
							<text class="iconfont icon-rate_review"></text>
						</view>
						<text class="mini-app-label">评教系统</text>
					</view>
				</view>
			</view>

			<!-- 校园生活 -->
			<view class="section card-section">
				<view class="section-header-inner">
					<view class="indicator bg-orange"></view>
					<text class="section-title-inner">校园生活</text>
				</view>
				<view class="mini-app-grid">
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-map"></text>
						</view>
						<text class="mini-app-label">校园导览</text>
					</view>
					<view class="mini-app-item" @tap="navigateTo('/pages/bus/index', 'bus')">
						<view class="mini-app-icon">
							<text class="iconfont icon-directions_bus"></text>
						</view>
						<text class="mini-app-label">校车时刻</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-wifi"></text>
						</view>
						<text class="mini-app-label">网络报修</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-sports_basketball"></text>
						</view>
						<text class="mini-app-label">场馆预约</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-local_laundry_service"></text>
						</view>
						<text class="mini-app-label">洗衣机</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-local_post_office"></text>
						</view>
						<text class="mini-app-label">快递查询</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-qr_code"></text>
						</view>
						<text class="mini-app-label">出入码</text>
					</view>
				</view>
			</view>

			<!-- 行政资讯 -->
			<view class="section card-section">
				<view class="section-header-inner">
					<view class="indicator bg-teal"></view>
					<text class="section-title-inner">行政资讯</text>
				</view>
				<view class="mini-app-grid">
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-campaign"></text>
						</view>
						<text class="mini-app-label">通知公告</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-account_balance"></text>
						</view>
						<text class="mini-app-label">学校概况</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-contact_phone"></text>
						</view>
						<text class="mini-app-label">黄页电话</text>
					</view>
					<view class="mini-app-item">
						<view class="mini-app-icon">
							<text class="iconfont icon-calendar_today"></text>
						</view>
						<text class="mini-app-label">校历</text>
					</view>
				</view>
			</view>

			<!-- Banner -->
			<view class="banner-section">
				<view class="banner-card">
					<image 
						class="banner-image" 
						src="https://images.unsplash.com/photo-1541339907198-e08756dedf3f?ixlib=rb-4.0.3&auto=format&fit=crop&w=1000&q=80" 
						mode="aspectFill"
					/>
					<view class="banner-overlay">
						<text class="banner-title">新生入学指南</text>
						<view class="banner-action">
							<text class="banner-desc">点击查看详细流程</text>
							<text class="iconfont icon-arrow_forward"></text>
						</view>
					</view>
				</view>
			</view>

			<!-- 底部占位 -->
			<view class="bottom-spacer"></view>
		</scroll-view>
		
		<!-- 自定义 TabBar -->
		<TabBar :current="1" />
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getUserInfo } from '@/services/apiService';
import { recordAppUsage as recordUsage, getAppUsage } from '@/utils/appUsage';
import TabBar from '@/components/TabBar/index.vue';

interface UserDisplay {
	avatar: string;
	name: string;
}

interface AppItem {
	id: string;
	label: string;
	icon: string;
	iconBg: string;
	path: string;
	lastUsed?: number;
}

// 所有可用的应用
const allApps: AppItem[] = [
	{ id: 'schedule', label: '课程表', icon: 'icon-calendar_today', iconBg: 'bg-blue', path: '/pages/home/index' },
	{ id: 'grades', label: '成绩查询', icon: 'icon-school', iconBg: 'bg-purple', path: '/pages/grades/index' },
	{ id: 'exams', label: '考试安排', icon: 'icon-fact_check', iconBg: 'bg-green', path: '/pages/exams/index' },
	{ id: 'evaluation', label: '评教系统', icon: 'icon-rate_review', iconBg: 'bg-indigo', path: '/pages/evaluation/index' },
	{ id: 'ai-chat', label: 'AI 助手', icon: 'icon-auto_awesome_mosaic', iconBg: 'bg-indigo', path: '/pages/ai-chat/index' },
	{ id: 'bus', label: '校车时刻', icon: 'icon-directions_bus', iconBg: 'bg-teal', path: '/pages/bus/index' },
];

const user = ref<UserDisplay>({
	avatar: '/static/default-avatar.png',
	name: '用户'
});

const recentApps = ref<AppItem[]>([]);

// 加载最近使用的应用
const loadRecentApps = () => {
	try {
		const recentUsage = getAppUsage();
		
		// 将使用记录与应用信息合并
		const appsWithUsage = allApps.map(app => ({
			...app,
			lastUsed: recentUsage[app.id] || 0
		}));
		
		// 按最后使用时间排序，取前4个
		recentApps.value = appsWithUsage
			.filter(app => app.lastUsed > 0)
			.sort((a, b) => (b.lastUsed || 0) - (a.lastUsed || 0))
			.slice(0, 4);
		
		// 如果不足4个，用默认应用补充
		if (recentApps.value.length < 4) {
			const usedIds = new Set(recentApps.value.map(app => app.id));
			const remaining = allApps.filter(app => !usedIds.has(app.id)).slice(0, 4 - recentApps.value.length);
			recentApps.value = [...recentApps.value, ...remaining];
		}
	} catch (error) {
		console.error('[Apps] Failed to load recent apps:', error);
		recentApps.value = allApps.slice(0, 4);
	}
};

onMounted(() => {
	const userInfo = getUserInfo<{ avatar?: string; realName?: string; username?: string }>();
	if (userInfo) {
		user.value = {
			avatar: userInfo.avatar || '/static/default-avatar.png',
			name: userInfo.realName || userInfo.username || '用户'
		};
	}
	
	loadRecentApps();
});

const handleSearch = () => {
	uni.showToast({ title: '搜索', icon: 'none' });
};

const navigateTo = (url: string, appId?: string) => {
	// 记录使用
	if (appId) {
		recordUsage(appId);
	}
	
	if (url.includes('/home/') || url.includes('/apps/') || url.includes('/profile/')) {
		uni.switchTab({ url });
	} else {
		uni.navigateTo({ url });
	}
};
</script>

<style lang="scss" scoped>
.apps-page {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	background-color: $bg-light;
}

.header {
	position: sticky;
	top: 0;
	z-index: 40;
	background: rgba(248, 250, 252, 0.8);
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
	flex-direction: column;
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

.header-right {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.search-btn {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	
	.icon {
		font-size: 36rpx;
	}
	
	&:active {
		background-color: rgba(0,0,0,0.05);
	}
}

.avatar {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	border: 2rpx solid $border-color;
}

.content {
	flex: 1;
	padding: 24rpx 32rpx;
}

.section {
	margin-bottom: 32rpx;
}

.section-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 24rpx;
}

.section-title {
	font-size: 36rpx;
	font-weight: 700;
	color: $text-primary;
}

.section-action {
	font-size: 24rpx;
	color: $primary;
	font-weight: 500;
}

.app-grid {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 24rpx;
}

.app-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
}

.app-icon {
	width: 112rpx;
	height: 112rpx;
	border-radius: 32rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.05);
	
	.icon {
		font-size: 40rpx;
	}
	
	&.bg-blue { background-color: #DBEAFE; }
	&.bg-orange { background-color: #FFEDD5; }
	&.bg-indigo { background-color: #E0E7FF; }
	&.bg-purple { background-color: #EDE9FE; }
	&.bg-green { background-color: #D1FAE5; }
	&.bg-teal { background-color: #CCFBF1; }
}

.app-label {
	font-size: 24rpx;
	color: $text-secondary;
	font-weight: 500;
	text-align: center;
}

.card-section {
	background-color: #fff;
	border-radius: 48rpx;
	padding: 40rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
}

.section-header-inner {
	display: flex;
	align-items: center;
	margin-bottom: 32rpx;
}

.indicator {
	width: 8rpx;
	height: 32rpx;
	border-radius: 4rpx;
	margin-right: 16rpx;
	
	&.bg-primary { background-color: $primary; }
	&.bg-orange { background-color: #f97316; }
	&.bg-teal { background-color: #14b8a6; }
}

.section-title-inner {
	font-size: 32rpx;
	font-weight: 700;
	color: $text-primary;
}

.mini-app-grid {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 32rpx 16rpx;
}

.mini-app-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16rpx;
}

.mini-app-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 50%;
	background-color: #f8fafc;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 2rpx solid $border-color;
	
	.icon {
		font-size: 36rpx;
	}
}

.mini-app-label {
	font-size: 24rpx;
	color: $text-primary;
	text-align: center;
}

.banner-section {
	margin-top: 16rpx;
}

.banner-card {
	position: relative;
	height: 220rpx;
	border-radius: 32rpx;
	overflow: hidden;
	box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.1);
}

.banner-image {
	width: 100%;
	height: 100%;
}

.banner-overlay {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background: linear-gradient(to right, rgba(30, 58, 138, 0.9), transparent);
	padding: 32rpx 48rpx;
	display: flex;
	flex-direction: column;
	justify-content: center;
}

.banner-title {
	font-size: 36rpx;
	font-weight: 700;
	color: #fff;
	margin-bottom: 8rpx;
}

.banner-action {
	display: flex;
	align-items: center;
}

.banner-desc {
	font-size: 24rpx;
	color: rgba(255,255,255,0.9);
}

.arrow {
	font-size: 24rpx;
	color: rgba(255,255,255,0.9);
	margin-left: 8rpx;
}

.bottom-spacer {
	height: 200rpx;
}
</style>
