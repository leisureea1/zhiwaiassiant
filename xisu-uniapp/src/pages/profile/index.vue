<template>
	<view class="profile-page">
		<!-- 头部 -->
		<view class="header">
			<text class="page-title">个人中心</text>
			<view class="notification-btn">
				<text class="iconfont icon-campaign"></text>
			</view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- 用户卡片 -->
			<view class="profile-card">
				<view class="avatar-wrapper">
					<image class="avatar" :src="user.avatar" mode="aspectFill" />
					<view class="online-dot"></view>
				</view>
				<view class="user-info">
					<view class="user-basic">
						<text class="user-name">{{ user.name }}</text>
						<text class="user-id">{{ user.studentId || '未绑定学号' }}</text>
					</view>
					<view class="user-tags">
						<text class="tag tag-blue" v-if="user.college">{{ user.college }}</text>
						<text class="tag tag-purple" v-if="user.major">{{ user.major }}</text>
						<text class="tag tag-green" v-if="user.className">{{ user.className }}</text>
					</view>
				</view>
			</view>

			<!-- 安全设置 -->
			<view class="section-card">
				<view class="menu-item" @tap="handleMenuItem('修改门户密码')">
					<view class="menu-icon bg-orange">
						<text class="iconfont icon-arrow_forward"></text>
					</view>
					<text class="menu-label">更新掌上西外密码</text>
					<text class="arrow">›</text>
				</view>
				<view class="divider"></view>
				<view class="menu-item" @tap="handleMenuItem('修改知外助手密码')">
					<view class="menu-icon bg-teal">
						<text class="iconfont icon-flight_class"></text>
					</view>
					<text class="menu-label">修改知外助手密码</text>
					<text class="arrow">›</text>
				</view>
			</view>

			<!-- 应用设置 -->
			<view class="section-card">
				<!-- #ifdef MP-WEIXIN -->
				<button
					class="menu-item menu-item-button"
					open-type="chooseAvatar"
					@chooseavatar="onChooseAvatar"
					@error="onChooseAvatarError"
				>
					<view class="menu-icon bg-blue">
						<text class="iconfont icon-account_circle"></text>
					</view>
					<text class="menu-label">更换头像</text>
					<text class="arrow">›</text>
				</button>
				<!-- #endif -->

				<!-- #ifndef MP-WEIXIN -->
				<view class="menu-item" @tap="handleAvatarUpload">
					<view class="menu-icon bg-blue">
						<text class="iconfont icon-account_circle"></text>
					</view>
					<text class="menu-label">更换头像</text>
					<text class="arrow">›</text>
				</view>
				<!-- #endif -->
				<view class="divider"></view>
				<view class="menu-item" @tap="handleMenuItem('设置')">
					<view class="menu-icon bg-gray">
						<text class="iconfont icon-settings"></text>
					</view>
					<text class="menu-label">设置</text>
					<text class="arrow">›</text>
				</view>
				<view class="divider"></view>
				<view class="menu-item" @tap="handleMenuItem('关于我们')">
					<view class="menu-icon bg-indigo">
						<text class="iconfont icon-school"></text>
					</view>
					<text class="menu-label">关于我们</text>
					<view class="version-tag">Leisure</view>
					<text class="arrow">›</text>
				</view>
			</view>

			<!-- 退出登录 -->
			<view class="logout-btn" @tap="handleLogout">
				<text class="logout-text">退出登录</text>
			</view>

			<!-- 底部占位 -->
			<view class="bottom-spacer"></view>
		</scroll-view>
		
		<!-- 隐私授权弹窗（微信小程序） -->
		<!-- #ifdef MP-WEIXIN -->
		<view v-if="showPrivacyDialog" class="privacy-mask">
			<view class="privacy-dialog">
				<view class="privacy-title">隐私授权提示</view>
				<view class="privacy-content">
					为了完成头像选择与上传，需要你先同意《小程序用户隐私保护指引》。
				</view>
				<view class="privacy-actions">
					<button class="privacy-btn privacy-btn-ghost" @tap="openPrivacyContract">
						查看指引
					</button>
					<button class="privacy-btn privacy-btn-outline" @tap="handleDisagreePrivacy">
						暂不授权
					</button>
					<button
						id="profile-privacy-agree-btn"
						class="privacy-btn privacy-btn-primary"
						open-type="agreePrivacyAuthorization"
						@agreeprivacyauthorization="handleAgreePrivacyAuthorization"
					>
						同意并继续
					</button>
				</view>
			</view>
		</view>
		<!-- #endif -->

		<!-- 自定义 TabBar -->
		<TabBar :current="2" />
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { getUserInfo, clearTokens, userApi, saveUserInfo, authApi, getRefreshToken } from '@/services/apiService';
import { PagePath } from '@/types';
import TabBar from '@/components/TabBar/index.vue';

declare const wx: any;

interface UserProfile {
	studentId: string;
	name: string;
	avatar: string;
	major: string;
	type: string;
	college?: string;
	className?: string;
}

const user = ref<UserProfile>({
	studentId: '',
	name: '加载中...',
	avatar: '/static/default-avatar.png',
	major: '',
	type: '本科生',
});

const isLoading = ref(true);
const backendUserId = ref('');
const showPrivacyDialog = ref(false);
const PRIVACY_AGREE_BUTTON_ID = 'profile-privacy-agree-btn';

type PrivacyResolve = (payload: { event: 'exposureAuthorization' | 'agree' | 'disagree'; buttonId?: string }) => void;
const pendingPrivacyResolve = ref<PrivacyResolve | null>(null);

// 从本地存储和后端加载用户信息
const loadUserInfo = async () => {
	isLoading.value = true;
	
	// 先从本地存储获取缓存的用户信息
	const cachedUser = getUserInfo<{
		id?: string;
		username?: string;
		realName?: string;
		studentId?: string;
		student_id?: string;
		avatar?: string;
		college?: string;
		major?: string;
		className?: string;
	}>();
	
	if (cachedUser) {
		backendUserId.value = cachedUser.id || '';
		const mappedStudentId =
			cachedUser.studentId ||
			cachedUser.student_id ||
			(cachedUser.username && /^\d{8,}$/.test(cachedUser.username) ? cachedUser.username : '') ||
			'';
		user.value = {
			studentId: mappedStudentId,
			name: cachedUser.realName || cachedUser.username || '未知用户',
			avatar: cachedUser.avatar || '/static/default-avatar.png',
			major: cachedUser.major || '',
			type: '本科生',
			college: cachedUser.college,
			className: cachedUser.className,
		};
	}
	
	// 尝试从后端刷新用户信息
	try {
		const freshUser = await userApi.getMe();
		console.log('[Profile] Fresh user data:', JSON.stringify(freshUser));
		if (freshUser) {
			backendUserId.value = freshUser.id || '';
			const mappedStudentId =
				freshUser.studentId ||
				((freshUser.username && /^\d{8,}$/.test(freshUser.username)) ? freshUser.username : '') ||
				'';
			user.value = {
				studentId: mappedStudentId,
				name: freshUser.realName || freshUser.username || '未知用户',
				avatar: freshUser.avatar || '/static/default-avatar.png',
				major: freshUser.major || '',
				type: '本科生',
				college: freshUser.college,
				className: freshUser.className,
			};
			console.log('[Profile] User value after update:', JSON.stringify(user.value));
			
			// 同时更新本地存储
			saveUserInfo(freshUser);
		}
	} catch (error) {
		console.log('[Profile] Failed to refresh user info:', error);
	} finally {
		isLoading.value = false;
	}
};

const uploadAvatarByFilePath = async (filePath: string) => {
	if (!filePath) {
		uni.showToast({ title: '未选择图片', icon: 'none' });
		return;
	}

	if (!backendUserId.value) {
		try {
			const freshUser = await userApi.getMe();
			if (freshUser?.id) {
				backendUserId.value = freshUser.id;
				saveUserInfo(freshUser);
			}
		} catch {
			uni.showToast({ title: '用户信息获取失败，请稍后重试', icon: 'none' });
			return;
		}

		if (!backendUserId.value) {
			uni.showToast({ title: '用户信息未加载完成', icon: 'none' });
			return;
		}
	}

	uni.showLoading({ title: '上传头像中...' });
	try {
		const uploaded = await userApi.uploadAvatarFile(backendUserId.value, filePath);
		if (uploaded?.avatar) {
			user.value.avatar = uploaded.avatar;
			const cached = getUserInfo<Record<string, unknown>>() || {};
			saveUserInfo({
				...cached,
				avatar: uploaded.avatar,
				id: backendUserId.value,
				studentId: user.value.studentId,
			});
		}
		uni.showToast({ title: '头像更新成功', icon: 'success' });
	} catch (error) {
		const message = error instanceof Error ? error.message : '头像上传失败';
		uni.showToast({ title: message, icon: 'none' });
	} finally {
		uni.hideLoading();
	}
};

const onChooseAvatar = async (e: { detail?: { avatarUrl?: string } }) => {
	const filePath = e?.detail?.avatarUrl || '';
	await uploadAvatarByFilePath(filePath);
};

const onChooseAvatarError = () => {
	uni.showToast({ title: '微信隐私声明未完成，改用相册上传', icon: 'none' });
	handleAvatarUpload();
};

const openPrivacyContract = () => {
	// #ifdef MP-WEIXIN
	if (typeof wx === 'undefined' || typeof wx.openPrivacyContract !== 'function') {
		uni.showToast({ title: '当前微信版本不支持', icon: 'none' });
		return;
	}

	wx.openPrivacyContract({
		success: () => {},
		fail: () => {
			uni.showToast({ title: '打开隐私指引失败', icon: 'none' });
		},
	});
	// #endif
};

const handleAgreePrivacyAuthorization = () => {
	pendingPrivacyResolve.value?.({
		event: 'agree',
		buttonId: PRIVACY_AGREE_BUTTON_ID,
	});
	pendingPrivacyResolve.value = null;
	showPrivacyDialog.value = false;
};

const handleDisagreePrivacy = () => {
	pendingPrivacyResolve.value?.({ event: 'disagree' });
	pendingPrivacyResolve.value = null;
	showPrivacyDialog.value = false;
	uni.showToast({ title: '你已拒绝隐私授权', icon: 'none' });
};

const handleAvatarUpload = async () => {
	uni.chooseImage({
		count: 1,
		sizeType: ['compressed'],
		sourceType: ['album', 'camera'],
		success: async (res) => {
			const filePath = res.tempFilePaths?.[0];
			await uploadAvatarByFilePath(filePath || '');
		},
		fail: (err) => {
			console.warn('[Profile] chooseImage failed:', err);
			uni.showToast({ title: '未能打开相册或相机', icon: 'none' });
		},
	});
};

onMounted(() => {
	// #ifdef MP-WEIXIN
	if (typeof wx !== 'undefined' && typeof wx.onNeedPrivacyAuthorization === 'function') {
		wx.onNeedPrivacyAuthorization((resolve: PrivacyResolve, eventInfo: { referrer?: string }) => {
			console.log('[Profile] Need privacy authorization from:', eventInfo?.referrer || 'unknown');
			pendingPrivacyResolve.value = resolve;
			showPrivacyDialog.value = true;
			resolve({ event: 'exposureAuthorization' });
		});
	}
	// #endif

	loadUserInfo();
});

onUnmounted(() => {
	pendingPrivacyResolve.value = null;
	showPrivacyDialog.value = false;
});

const handleMenuItem = (name: string) => {
	switch (name) {
		case '修改门户密码':
			uni.navigateTo({ url: '/pages/profile/update-jwxt-password' });
			break;
		case '修改知外助手密码':
			uni.navigateTo({ url: '/pages/profile/change-password' });
			break;
		case '关于我们':
			uni.navigateTo({ url: '/pages/profile/about' });
			break;
		default:
			uni.showToast({ title: name, icon: 'none' });
	}
};

const handleLogout = () => {
	uni.showModal({
		title: '提示',
		content: '确定要退出登录吗？',
		success: async (res) => {
			if (res.confirm) {
				const refreshToken = getRefreshToken() || undefined;
				await authApi.logout(refreshToken).catch(() => {});
				clearTokens();
				uni.reLaunch({
					url: PagePath.LOGIN
				});
			}
		}
	});
};
</script>

<style lang="scss" scoped>
.profile-page {
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
	justify-content: space-between;
	align-items: center;
	padding: 24rpx 48rpx;
	padding-top: calc(var(--status-bar-height) + 48rpx);
	background: rgba(248, 250, 252, 0.9);
	backdrop-filter: blur(20rpx);
}

.page-title {
	font-size: 40rpx;
	font-weight: 700;
	color: $text-primary;
}

.notification-btn {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	background-color: #fff;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.05);
	display: flex;
	align-items: center;
	justify-content: center;
	
	.icon {
		font-size: 36rpx;
	}
}

.content {
	flex: 1;
	padding: 0 32rpx;
}

.profile-card {
	display: flex;
	align-items: center;
	background-color: #fff;
	border-radius: 48rpx;
	padding: 40rpx 48rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
}

.avatar-wrapper {
	position: relative;
	margin-right: 40rpx;
}

.avatar {
	width: 160rpx;
	height: 160rpx;
	border-radius: 50%;
	background: #f8fafc;
	border: 2rpx solid #e2e8f0;
	box-shadow: 0 6rpx 18rpx rgba(15, 23, 42, 0.08);
}

.online-dot {
	position: absolute;
	bottom: 8rpx;
	right: 8rpx;
	width: 24rpx;
	height: 24rpx;
	background-color: $success;
	border: 4rpx solid #fff;
	border-radius: 50%;
}

.user-info {
	flex: 1;
}

.user-basic {
	display: flex;
	align-items: baseline;
	gap: 16rpx;
	margin-bottom: 16rpx;
	flex-wrap: wrap;
}

.user-name {
	font-size: 40rpx;
	font-weight: 700;
	color: $text-primary;
}

.user-id {
	font-size: 28rpx;
	color: $text-secondary;
}

.user-tags {
	display: flex;
	flex-wrap: wrap;
	gap: 12rpx;
}

.tag {
	display: inline-block;
	padding: 8rpx 20rpx;
	border-radius: 32rpx;
	font-size: 24rpx;
	font-weight: 500;
	
	&.tag-blue {
		background-color: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}
	
	&.tag-purple {
		background-color: rgba(139, 92, 246, 0.1);
		color: #8b5cf6;
	}
	
	&.tag-green {
		background-color: rgba(16, 185, 129, 0.1);
		color: #10b981;
	}
}

.section-card {
	background-color: #fff;
	border-radius: 48rpx;
	overflow: hidden;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
}

.menu-item {
	display: flex;
	align-items: center;
	padding: 32rpx;
	
	&:active {
		background-color: rgba(0,0,0,0.02);
	}
}

.menu-item-button {
	width: 100%;
	text-align: left;
	background: #fff;
	border: none;
	border-radius: 0;
	line-height: normal;

	&::after {
		border: none;
	}
}

.privacy-mask {
	position: fixed;
	left: 0;
	top: 0;
	right: 0;
	bottom: 0;
	z-index: 999;
	background: rgba(15, 23, 42, 0.45);
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 32rpx;
}

.privacy-dialog {
	width: 100%;
	max-width: 620rpx;
	background: #fff;
	border-radius: 28rpx;
	padding: 32rpx;
	box-shadow: 0 20rpx 60rpx rgba(15, 23, 42, 0.16);
}

.privacy-title {
	font-size: 34rpx;
	font-weight: 700;
	color: #0f172a;
	margin-bottom: 16rpx;
}

.privacy-content {
	font-size: 28rpx;
	line-height: 1.6;
	color: #334155;
	margin-bottom: 28rpx;
}

.privacy-actions {
	display: flex;
	gap: 16rpx;
}

.privacy-btn {
	flex: 1;
	height: 76rpx;
	border-radius: 16rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 28rpx;
	line-height: normal;
}

.privacy-btn-ghost {
	background: #f8fafc;
	color: #0f172a;
	border: 1rpx solid #e2e8f0;
}

.privacy-btn-outline {
	background: #fff;
	color: #334155;
	border: 1rpx solid #cbd5e1;
}

.privacy-btn-primary {
	background: #2563eb;
	color: #fff;
	border: 1rpx solid #2563eb;
}

.menu-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-right: 24rpx;
	
	.icon {
		font-size: 36rpx;
	}
	
	&.bg-orange { background-color: rgba(249, 115, 22, 0.1); }
	&.bg-teal { background-color: rgba(20, 184, 166, 0.1); }
	&.bg-blue { background-color: rgba(59, 130, 246, 0.1); }
	&.bg-gray { background-color: rgba(107, 114, 128, 0.1); }
	&.bg-indigo { background-color: rgba(99, 102, 241, 0.1); }
}

.menu-label {
	flex: 1;
	font-size: 32rpx;
	font-weight: 500;
	color: $text-primary;
}

.version-tag {
	padding: 8rpx 16rpx;
	background-color: rgba(0,0,0,0.04);
	border-radius: 12rpx;
	font-size: 24rpx;
	color: $text-light;
	margin-right: 8rpx;
}

.arrow {
	font-size: 40rpx;
	color: $text-light;
}

.divider {
	height: 2rpx;
	background-color: $border-color;
	margin-left: 136rpx;
}

.logout-btn {
	background-color: rgba(239, 68, 68, 0.05);
	border-radius: 32rpx;
	padding: 32rpx;
	margin-top: 16rpx;
	box-shadow: 0 2rpx 12rpx rgba(239, 68, 68, 0.05);
	
	&:active {
		background-color: rgba(239, 68, 68, 0.1);
	}
}

.logout-text {
	display: block;
	text-align: center;
	font-size: 32rpx;
	font-weight: 700;
	color: $danger;
}

.bottom-spacer {
	height: 200rpx;
}
</style>
