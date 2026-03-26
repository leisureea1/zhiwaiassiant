<template>
	<view class="page">
		<!-- 头部 -->
		<view class="header">
			<view class="back-btn" @tap="handleBack">
				<text class="iconfont icon-navigate_before"></text>
			</view>
			<text class="page-title">更新掌上西外密码</text>
			<view class="header-right"></view>
		</view>

		<!-- 内容区域 -->
		<scroll-view class="content" scroll-y>
			<view class="info-box">
				<view class="info-icon bg-orange">
					<text class="iconfont icon-arrow_forward"></text>
				</view>
				<view class="info-body">
					<text class="info-title">同步掌上西外密码</text>
					<text class="info-text">如果您的掌上西外密码已更改，请在此处提交最新密码，以便继续使用成绩查询、考试安排等功能。</text>
				</view>
			</view>

			<view class="form-section">
				<view class="input-wrapper">
					<view class="input-label">学号</view>
					<view class="input-box">
						<input 
							class="input-field input-disabled" 
							type="text" 
							v-model="username"
							placeholder="请输入学号" 
							placeholder-class="placeholder"
							disabled
						/>
					</view>
				</view>

				<view class="input-wrapper">
					<view class="input-label">新掌上西外密码</view>
					<view class="input-box">
						<input 
							class="input-field" 
							type="text"
							:password="!showPassword"
							v-model="password"
							placeholder="请输入最新的掌上西外密码" 
							placeholder-class="placeholder"
						/>
						<view class="toggle-password" @tap="showPassword = !showPassword">
							<text class="iconfont" :class="showPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
				</view>
			</view>

			<view class="tip-card">
				<text class="iconfont icon-info"></text>
				<text class="tip-content">请输入掌上西外的最新登录密码</text>
			</view>

			<button class="btn-primary" :disabled="isLoading" @tap="handleSubmit">
				{{ isLoading ? '提交中...' : '更新密码' }}
			</button>
		</scroll-view>
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { jwxtApi, getUserInfo } from '@/services/apiService';

const username = ref('');
const password = ref('');
const showPassword = ref(false);
const isLoading = ref(false);

onMounted(() => {
	// 预填充学号
	const userInfo = getUserInfo<{ studentId?: string }>();
	if (userInfo?.studentId) {
		username.value = userInfo.studentId;
	}
});

const handleBack = () => {
	safeNavigateBack();
};

const handleSubmit = async () => {
	if (!username.value) {
		uni.showToast({ title: '请输入学号', icon: 'none' });
		return;
	}
	if (!password.value) {
		uni.showToast({ title: '请输入新密码', icon: 'none' });
		return;
	}

	isLoading.value = true;
	uni.showLoading({ title: '验证中...' });

	try {
		await jwxtApi.bindAccount(username.value, password.value);
		uni.hideLoading();
		uni.showToast({ title: '掌上西外密码更新成功', icon: 'success' });
		setTimeout(() => {
			safeNavigateBack();
		}, 1500);
	} catch (error) {
		uni.hideLoading();
		const errorMsg = error instanceof Error ? error.message : '更新失败，请检查密码是否正确';
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
		color: #f97316;
	}

	&.bg-orange {
		background-color: rgba(249, 115, 22, 0.1);
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

.input-disabled {
	opacity: 0.6;
	padding-right: 32rpx;
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
	background-color: rgba(249, 115, 22, 0.06);
	border-radius: 32rpx;
	padding: 24rpx 32rpx;
	margin-bottom: 48rpx;
	
	.iconfont {
		font-size: 28rpx;
		color: #f97316;
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
	background-color: #f97316;
	border-radius: 48rpx;
	border: none;
	color: #fff;
	font-size: 32rpx;
	font-weight: 600;
	box-shadow: 0 8rpx 24rpx rgba(249, 115, 22, 0.25);

	&[disabled] {
		opacity: 0.5;
	}
}
</style>
