<template>
	<view class="login-page">
		<!-- 点阵背景 -->
		<view class="dot-pattern"></view>

		<view class="login-content">
			<!-- Logo区域 -->
			<view class="logo-section">
				<view class="logo-box">
					<view class="logo-inner">
						<text class="iconfont icon-school"></text>
					</view>
				</view>
				<view class="logo-text">
					<text class="app-title">知外助手</text>
					<text class="app-subtitle">欢迎回来，请登录继续</text>
				</view>
			</view>

			<!-- 升级提示 -->
			<view class="upgrade-notice">
				<text class="iconfont icon-info" style="font-size: 28rpx; color: #f59e0b; margin-right: 12rpx;"></text>
				<text class="notice-text">服务器升级，原账户数据已重置，请重新注册。</text>
			</view>

			<!-- 表单区域 -->
			<view class="form-section">
				<view class="input-group">
					<view class="input-wrapper">
						<view class="input-icon">
							<text class="iconfont icon-assignment_ind"></text>
						</view>
						<input 
							class="input-field" 
							type="text" 
							v-model="studentId"
							placeholder="学号/用户名/邮箱" 
							placeholder-class="placeholder"
						/>
					</view>

					<view class="input-wrapper">
						<view class="input-icon">
							<text class="iconfont icon-flight_class"></text>
						</view>
						<input 
							class="input-field" 
							type="text"
							:password="!showPassword"
							v-model="password"
							placeholder="密码" 
							placeholder-class="placeholder"
						/>
						<view class="toggle-password" @tap="showPassword = !showPassword">
							<text class="iconfont" :class="showPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>

					<view class="forgot-password">
						<text class="forgot-link" @tap="handleForgotPassword">忘记密码?</text>
					</view>
				</view>

				<view class="button-group">
					<button class="btn-primary" @tap="handleLogin">
						登录
					</button>
					<button class="btn-outline" @tap="handleRegister">
						注册
					</button>
				</view>
			</view>

			<!-- 底部协议 -->
			<view class="footer-agreement">
				<view class="agreement-row" @tap="agreedPrivacy = !agreedPrivacy">
					<view class="checkbox" :class="{ checked: agreedPrivacy }">
						<text v-if="agreedPrivacy" class="iconfont icon-check_circle" style="font-size: 36rpx; color: #3B82F6;"></text>
						<view v-else class="checkbox-empty"></view>
					</view>
					<text class="agreement-text">
						我已阅读并同意
					</text>
				</view>
				<view class="agreement-links">
					<text class="link" @tap.stop="openServiceAgreement">《用户服务协议》</text>
					<text class="agreement-text">和</text>
					<text class="link" @tap.stop="openPrivacyPolicy">《隐私政策》</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { authApi, saveTokens, saveUserInfo, getAccessToken, getUserInfo } from '@/services/apiService';

const studentId = ref('');
const password = ref('');
const showPassword = ref(false);
const isLoading = ref(false);
const agreedPrivacy = ref(false);

// 检查是否已登录
onMounted(() => {
	const token = getAccessToken();
	const user = getUserInfo();
	if (token && user) {
		// 已登录，跳转到首页
		uni.switchTab({ url: '/pages/home/index' });
	}
});

const openServiceAgreement = () => {
	uni.navigateTo({ url: '/pages/service-agreement/index' });
};

const openPrivacyPolicy = () => {
	uni.navigateTo({ url: '/pages/privacy-policy/index' });
};

const handleLogin = async () => {
	if (!agreedPrivacy.value) {
		uni.showToast({ title: '请先阅读并同意用户服务协议和隐私政策', icon: 'none', duration: 2000 });
		return;
	}
	if (!studentId.value) {
		uni.showToast({ title: '请输入学号/用户名/邮箱', icon: 'none' });
		return;
	}
	if (!password.value) {
		uni.showToast({ title: '请输入密码', icon: 'none' });
		return;
	}
	
	isLoading.value = true;
	uni.showLoading({ title: '登录中...' });
	
	try {
		const res = await authApi.login({
			studentId: studentId.value,
			password: password.value,
		});
		
		// 保存登录凭证和用户信息
		saveTokens({
			accessToken: res.accessToken,
			refreshToken: res.refreshToken,
		});
		saveUserInfo(res.user);
		
		uni.hideLoading();
		uni.showToast({ title: '登录成功', icon: 'success' });
		
		setTimeout(() => {
			uni.switchTab({ url: '/pages/home/index' });
		}, 1000);
	} catch (error) {
		uni.hideLoading();
		const errorMessage = error instanceof Error ? error.message : '登录失败';
		uni.showToast({ title: errorMessage, icon: 'none', duration: 2000 });
	} finally {
		isLoading.value = false;
	}
};

const handleRegister = () => {
	uni.navigateTo({
		url: '/pages/register/index'
	});
};

const handleForgotPassword = () => {
	uni.navigateTo({
		url: '/pages/forgot-password/index'
	});
};
</script>

<style lang="scss" scoped>
.login-page {
	position: relative;
	min-height: 100vh;
	background-color: $bg-light;
	overflow: hidden;
}

.dot-pattern {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	opacity: 0.4;
	background-image: radial-gradient(#9ca3af 2rpx, transparent 2rpx);
	background-size: 48rpx 48rpx;
	pointer-events: none;
}

.login-content {
	position: relative;
	z-index: 10;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	min-height: 100vh;
	padding: 40rpx 48rpx;	padding-top: calc(var(--status-bar-height) + 120rpx);	box-sizing: border-box;
}

.logo-section {
	display: flex;
	flex-direction: column;
	align-items: center;
	margin-bottom: 80rpx;
}

.logo-box {
	width: 160rpx;
	height: 160rpx;
	border-radius: 32rpx;
	background-color: #fff;
	box-shadow: 0 20rpx 40rpx rgba(0,0,0,0.1);
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 16rpx;
	margin-bottom: 32rpx;
}

.logo-inner {
	width: 100%;
	height: 100%;
	border-radius: 24rpx;
	background: linear-gradient(135deg, $primary 0%, #60A5FA 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	
	.iconfont {
		font-size: 64rpx;
		color: #fff;
	}
}

.logo-text {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.app-title {
	font-size: 48rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 8rpx;
}

.app-subtitle {
	font-size: 28rpx;
	color: $text-secondary;
}

.form-section {
	width: 100%;
}

.input-group {
	margin-bottom: 40rpx;
}

.input-wrapper {
	position: relative;
	margin-bottom: 32rpx;
	
	&:last-child {
		margin-bottom: 0;
	}
}

.input-icon {
	position: absolute;
	left: 32rpx;
	top: 50%;
	transform: translateY(-50%);
	
	.iconfont {
		font-size: 40rpx;
		color: $text-light;
	}
}

.input-field {
	width: 100%;
	height: 112rpx;
	border-radius: 24rpx;
	border: 2rpx solid $border-color;
	background-color: #fff;
	padding-left: 96rpx;
	padding-right: 96rpx;
	font-size: 32rpx;
	color: $text-primary;
	box-sizing: border-box;
	
	&:focus {
		border-color: $primary;
	}
}

.placeholder {
	color: $text-light;
}

.toggle-password {
	position: absolute;
	right: 0;
	top: 0;
	width: 96rpx;
	height: 112rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	
	.iconfont {
		font-size: 40rpx;
		color: $text-light;
	}
}

.forgot-password {
	display: flex;
	justify-content: flex-end;
	margin-top: 16rpx;
}

.forgot-link {
	font-size: 28rpx;
	color: $text-secondary;
}

.button-group {
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.btn-primary {
	width: 100%;
	height: 112rpx;
	border-radius: 24rpx;
	background-color: $primary;
	color: #fff;
	font-size: 32rpx;
	font-weight: 700;
	display: flex;
	align-items: center;
	justify-content: center;
	border: none;
	box-shadow: 0 16rpx 32rpx rgba(59, 130, 246, 0.3);
	
	&:active {
		transform: scale(0.98);
		opacity: 0.9;
	}
}

.btn-outline {
	width: 100%;
	height: 112rpx;
	border-radius: 24rpx;
	background-color: transparent;
	color: $primary;
	font-size: 32rpx;
	font-weight: 700;
	display: flex;
	align-items: center;
	justify-content: center;
	border: 2rpx solid $primary;
	
	&:active {
		background-color: rgba(59, 130, 246, 0.05);
	}
}

.footer-agreement {
	margin-top: auto;
	padding-top: 60rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
}

.agreement-row {
	display: flex;
	align-items: center;
	margin-bottom: 8rpx;
}

.checkbox {
	margin-right: 12rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.checkbox-empty {
	width: 32rpx;
	height: 32rpx;
	border-radius: 50%;
	border: 2rpx solid #ccc;
	background-color: #fff;
}

.agreement-text {
	font-size: 24rpx;
	color: $text-light;
}

.agreement-links {
	display: flex;
	align-items: center;
}

.link {
	font-size: 24rpx;
	color: $primary;
}
</style>
