<template>
	<view class="page">
		<!-- 头部 -->
		<view class="header">
			<view class="back-btn" @tap="handleBack">
				<text class="iconfont icon-navigate_before"></text>
			</view>
			<text class="page-title">修改知外助手密码</text>
			<view class="header-right"></view>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="content" scroll-y>
			<view class="info-box">
				<view class="info-icon bg-teal">
					<text class="iconfont icon-flight_class"></text>
				</view>
				<view class="info-body">
					<text class="info-title">修改登录密码</text>
					<text class="info-text">修改知外助手的登录密码。修改成功后需要重新登录。</text>
				</view>
			</view>

			<view class="form-section">
				<view class="input-wrapper">
					<view class="input-label">当前密码</view>
					<view class="input-box">
						<input 
							class="input-field" 
							type="text"
							:password="!showOldPassword"
							v-model="oldPassword"
							placeholder="请输入当前密码" 
							placeholder-class="placeholder"
						/>
						<view class="toggle-password" @tap="showOldPassword = !showOldPassword">
							<text class="iconfont" :class="showOldPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
				</view>

				<view class="input-wrapper">
					<view class="input-label">新密码</view>
					<view class="input-box">
						<input 
							class="input-field" 
							type="text"
							:password="!showNewPassword"
							v-model="newPassword"
							placeholder="请输入新密码（至少6位）" 
							placeholder-class="placeholder"
						/>
						<view class="toggle-password" @tap="showNewPassword = !showNewPassword">
							<text class="iconfont" :class="showNewPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
				</view>

				<view class="input-wrapper">
					<view class="input-label">确认新密码</view>
					<view class="input-box">
						<input 
							class="input-field" 
							type="text"
							:password="!showConfirmPassword"
							v-model="confirmPassword"
							placeholder="请再次输入新密码" 
							placeholder-class="placeholder"
						/>
						<view class="toggle-password" @tap="showConfirmPassword = !showConfirmPassword">
							<text class="iconfont" :class="showConfirmPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
				</view>
			</view>

			<view class="tip-card">
				<text class="iconfont icon-info"></text>
				<text class="tip-content">密码修改成功后，所有设备将自动退出登录</text>
			</view>

			<button class="btn-primary" :disabled="isLoading" @tap="handleSubmit">
				{{ isLoading ? '提交中...' : '确认修改' }}
			</button>
		</scroll-view>
	</view>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { authApi, clearTokens } from '@/services/apiService';

const oldPassword = ref('');
const newPassword = ref('');
const confirmPassword = ref('');
const showOldPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);
const isLoading = ref(false);

const handleBack = () => {
	safeNavigateBack();
};

const handleSubmit = async () => {
	if (!oldPassword.value) {
		uni.showToast({ title: '请输入当前密码', icon: 'none' });
		return;
	}
	if (!newPassword.value) {
		uni.showToast({ title: '请输入新密码', icon: 'none' });
		return;
	}
	if (newPassword.value.length < 6) {
		uni.showToast({ title: '新密码至少6位', icon: 'none' });
		return;
	}
	if (newPassword.value !== confirmPassword.value) {
		uni.showToast({ title: '两次输入的密码不一致', icon: 'none' });
		return;
	}
	if (oldPassword.value === newPassword.value) {
		uni.showToast({ title: '新密码不能与当前密码相同', icon: 'none' });
		return;
	}

	isLoading.value = true;
	uni.showLoading({ title: '修改中...' });

	try {
		await authApi.changePassword(oldPassword.value, newPassword.value);
		uni.hideLoading();
		
		uni.showModal({
			title: '修改成功',
			content: '密码已修改，请重新登录',
			showCancel: false,
			success: async () => {
				await authApi.logout().catch(() => {});
				clearTokens();
				uni.reLaunch({ url: '/pages/login/index' });
			}
		});
	} catch (error) {
		uni.hideLoading();
		const errorMsg = error instanceof Error ? error.message : '修改失败';
		uni.showToast({ title: errorMsg, icon: 'none', duration: 2000 });
	} finally {
		isLoading.value = false;
	}
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
	padding-top: 24rpx;
}

.info-box {
	display: flex;
	align-items: center;
	gap: 24rpx;
	background-color: #fff;
	border-radius: 48rpx;
	padding: 40rpx 48rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
	border: 2rpx solid rgba(0, 0, 0, 0.04);
}

.info-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;

	.iconfont {
		font-size: 36rpx;
		color: #14b8a6;
	}

	&.bg-teal {
		background-color: rgba(20, 184, 166, 0.1);
	}
}

.info-body {
	flex: 1;
}

.info-title {
	display: block;
	font-size: 32rpx;
	font-weight: 600;
	color: $text-primary;
	margin-bottom: 8rpx;
}

.info-text {
	font-size: 26rpx;
	color: $text-secondary;
	line-height: 1.6;
}

.form-section {
	background-color: #fff;
	border-radius: 48rpx;
	padding: 40rpx 48rpx;
	margin-bottom: 24rpx;
	box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.04);
	border: 2rpx solid rgba(0, 0, 0, 0.04);
}

.input-wrapper {
	margin-bottom: 32rpx;

	&:last-child {
		margin-bottom: 0;
	}
}

.input-label {
	font-size: 28rpx;
	color: $text-primary;
	font-weight: 500;
	margin-bottom: 16rpx;
}

.input-box {
	position: relative;
	display: flex;
	align-items: center;
}

.input-field {
	flex: 1;
	height: 96rpx;
	background: $bg-light;
	border-radius: 24rpx;
	padding: 0 96rpx 0 32rpx;
	font-size: 30rpx;
	color: $text-primary;
	box-sizing: border-box;
}

.placeholder {
	color: $text-light;
}

.toggle-password {
	position: absolute;
	right: 24rpx;
	top: 50%;
	transform: translateY(-50%);

	.iconfont {
		font-size: 36rpx;
		color: $text-light;
	}
}

.tip-card {
	display: flex;
	align-items: center;
	gap: 12rpx;
	background-color: rgba(20, 184, 166, 0.06);
	border-radius: 32rpx;
	padding: 24rpx 32rpx;
	margin-bottom: 48rpx;
	
	.iconfont {
		font-size: 28rpx;
		color: #14b8a6;
	}
	
	.tip-content {
		font-size: 24rpx;
		color: $text-secondary;
		line-height: 1.5;
	}
}

.btn-primary {
	width: 100%;
	height: 96rpx;
	line-height: 96rpx;
	background-color: #14b8a6;
	border-radius: 48rpx;
	border: none;
	color: #fff;
	font-size: 32rpx;
	font-weight: 600;
	box-shadow: 0 8rpx 24rpx rgba(20, 184, 166, 0.25);

	&[disabled] {
		opacity: 0.5;
	}
}
</style>
