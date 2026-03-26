<template>
	<view class="forgot-page">
		<!-- 点阵背景 -->
		<view class="dot-pattern"></view>

		<view class="forgot-content">
			<!-- 返回按钮 -->
			<view class="back-btn" @tap="handleBack">
				<text class="iconfont icon-arrow_back"></text>
			</view>

			<!-- 步骤指示器 -->
			<view class="step-indicator">
				<view 
					v-for="i in 3" 
					:key="i" 
					class="step-dot"
					:class="{ active: currentStep >= i, current: currentStep === i }"
				></view>
			</view>

			<!-- 步骤1: 输入邮箱 -->
			<view v-if="currentStep === 1" class="step-content">
				<view class="step-header">
					<text class="step-title">找回密码</text>
					<text class="step-subtitle">请输入您注册时使用的邮箱</text>
				</view>

				<view class="form-section">
					<view class="input-wrapper" :class="{ 'input-error': emailError }">
						<view class="input-icon">
							<text class="iconfont icon-email"></text>
						</view>
						<input 
							class="input-field" 
							type="text" 
							v-model="email"
							placeholder="请输入邮箱地址" 
							placeholder-class="placeholder"
							@blur="validateEmail"
						/>
					</view>
					<text v-if="emailError" class="error-hint">{{ emailError }}</text>

					<button 
						class="btn-primary" 
						@tap="sendCode"
						:disabled="!email || !!emailError || sending"
					>
						{{ sending ? '发送中...' : '发送验证码' }}
					</button>
				</view>
			</view>

			<!-- 步骤2: 输入验证码 -->
			<view v-if="currentStep === 2" class="step-content">
				<view class="step-header">
					<text class="step-title">验证邮箱</text>
					<text class="step-subtitle">验证码已发送至 {{ email }}</text>
				</view>

				<view class="form-section">
					<view class="input-wrapper" :class="{ 'input-error': codeError }">
						<view class="input-icon">
							<text class="iconfont icon-verified_user"></text>
						</view>
						<input 
							class="input-field" 
							type="number" 
							v-model="verificationCode"
							placeholder="请输入6位验证码" 
							placeholder-class="placeholder"
							maxlength="6"
						/>
						<button 
							class="resend-btn"
							:class="{ disabled: countdown > 0 }"
							:disabled="countdown > 0"
							@tap="resendCode"
						>
							{{ countdown > 0 ? `${countdown}s` : '重新发送' }}
						</button>
					</view>
					<text v-if="codeError" class="error-hint">{{ codeError }}</text>

					<button 
						class="btn-primary" 
						@tap="goToStep3"
						:disabled="verificationCode.length !== 6"
					>
						下一步
					</button>
				</view>
			</view>

			<!-- 步骤3: 设置新密码 -->
			<view v-if="currentStep === 3" class="step-content">
				<view class="step-header">
					<text class="step-title">设置新密码</text>
					<text class="step-subtitle">请设置您的新密码</text>
				</view>

				<view class="form-section">
					<view class="input-wrapper" :class="{ 'input-error': passwordError }">
						<view class="input-icon">
							<text class="iconfont icon-flight_class"></text>
						</view>
						<input 
							class="input-field" 
							type="text"
							:password="!showPassword"
							v-model="newPassword"
							placeholder="设置新密码 (至少8位)" 
							placeholder-class="placeholder"
							@input="validatePassword"
						/>
						<view class="toggle-password" @tap="showPassword = !showPassword">
							<text class="iconfont" :class="showPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
					<text v-if="passwordError" class="error-hint">{{ passwordError }}</text>

					<!-- 密码强度指示器 -->
					<view v-if="newPassword && !passwordError" class="password-strength">
						<view class="strength-bars">
							<view class="strength-bar" :class="{ active: passwordStrength >= 1 }" :style="{ backgroundColor: passwordStrength >= 1 ? strengthColor : '' }"></view>
							<view class="strength-bar" :class="{ active: passwordStrength >= 2 }" :style="{ backgroundColor: passwordStrength >= 2 ? strengthColor : '' }"></view>
							<view class="strength-bar" :class="{ active: passwordStrength >= 3 }" :style="{ backgroundColor: passwordStrength >= 3 ? strengthColor : '' }"></view>
							<view class="strength-bar" :class="{ active: passwordStrength >= 4 }" :style="{ backgroundColor: passwordStrength >= 4 ? strengthColor : '' }"></view>
						</view>
						<text class="strength-text" :class="strengthClass">{{ strengthText }}</text>
					</view>

					<view class="input-wrapper" :class="{ 'input-error': confirmPasswordError }">
						<view class="input-icon">
							<text class="iconfont icon-flight_class"></text>
						</view>
						<input 
							class="input-field" 
							type="text"
							:password="!showConfirmPassword"
							v-model="confirmPassword"
							placeholder="确认新密码" 
							placeholder-class="placeholder"
							@input="validateConfirmPassword"
						/>
						<view class="toggle-password" @tap="showConfirmPassword = !showConfirmPassword">
							<text class="iconfont" :class="showConfirmPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
					<text v-if="confirmPasswordError" class="error-hint">{{ confirmPasswordError }}</text>

					<button 
						class="btn-primary" 
						@tap="resetPassword"
						:disabled="!canSubmit || submitting"
					>
						{{ submitting ? '重置中...' : '重置密码' }}
					</button>
				</view>
			</view>

			<!-- 成功提示 -->
			<view v-if="currentStep === 4" class="step-content success-content">
				<view class="success-icon-box">
					<text class="iconfont icon-check_circle"></text>
				</view>
				<text class="success-title">密码重置成功！</text>
				<text class="success-subtitle">请使用新密码登录</text>
				<button class="btn-primary" @tap="goToLogin">
					返回登录
				</button>
			</view>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { authApi } from '@/services/apiService';

// 步骤控制
const currentStep = ref(1);

// 表单数据
const email = ref('');
const verificationCode = ref('');
const resetToken = ref('');
const newPassword = ref('');
const confirmPassword = ref('');

// 显示密码控制
const showPassword = ref(false);
const showConfirmPassword = ref(false);

// 错误信息
const emailError = ref('');
const codeError = ref('');
const passwordError = ref('');
const confirmPasswordError = ref('');

// 状态控制
const sending = ref(false);
const submitting = ref(false);
const countdown = ref(0);
let countdownTimer: ReturnType<typeof setInterval> | null = null;

// 邮箱验证
const validateEmail = () => {
	if (!email.value) {
		emailError.value = '请输入邮箱地址';
		return false;
	}
	const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	if (!emailRegex.test(email.value)) {
		emailError.value = '请输入有效的邮箱地址';
		return false;
	}
	emailError.value = '';
	return true;
};

// 密码验证
const validatePassword = () => {
	if (!newPassword.value) {
		passwordError.value = '';
		return false;
	}
	if (newPassword.value.length < 8) {
		passwordError.value = '密码至少8个字符';
		return false;
	}
	passwordError.value = '';
	validateConfirmPassword();
	return true;
};

// 确认密码验证
const validateConfirmPassword = () => {
	if (!confirmPassword.value) {
		confirmPasswordError.value = '';
		return false;
	}
	if (confirmPassword.value !== newPassword.value) {
		confirmPasswordError.value = '两次输入的密码不一致';
		return false;
	}
	confirmPasswordError.value = '';
	return true;
};

// 密码强度
const passwordStrength = computed(() => {
	const pwd = newPassword.value;
	if (!pwd) return 0;
	let strength = 0;
	if (pwd.length >= 8) strength++;
	if (pwd.length >= 10) strength++;
	if (/[A-Z]/.test(pwd) && /[a-z]/.test(pwd)) strength++;
	if (/[0-9]/.test(pwd) && /[^A-Za-z0-9]/.test(pwd)) strength++;
	return strength;
});

const strengthText = computed(() => {
	const texts = ['', '弱', '中', '强', '很强'];
	return texts[passwordStrength.value];
});

const strengthColor = computed(() => {
	const colors = ['', '#ef4444', '#f59e0b', '#10b981', '#3b82f6'];
	return colors[passwordStrength.value];
});

const strengthClass = computed(() => {
	const classes = ['', 'weak', 'medium', 'strong', 'very-strong'];
	return classes[passwordStrength.value];
});

// 是否可以提交
const canSubmit = computed(() => {
	return newPassword.value.length >= 8 && 
		confirmPassword.value === newPassword.value &&
		!passwordError.value &&
		!confirmPasswordError.value;
});

// 发送验证码
const sendCode = async () => {
	if (!validateEmail()) return;
	
	sending.value = true;
	try {
		const result = await authApi.forgotPassword(email.value);
		resetToken.value = result?.token || '';
		uni.showToast({ title: '验证码已发送', icon: 'success' });
		currentStep.value = 2;
		startCountdown();
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : '发送失败';
		uni.showToast({ title: errorMessage, icon: 'none' });
	} finally {
		sending.value = false;
	}
};

// 重新发送验证码
const resendCode = async () => {
	if (countdown.value > 0) return;
	
	sending.value = true;
	try {
		const result = await authApi.forgotPassword(email.value);
		resetToken.value = result?.token || '';
		uni.showToast({ title: '验证码已发送', icon: 'success' });
		startCountdown();
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : '发送失败';
		uni.showToast({ title: errorMessage, icon: 'none' });
	} finally {
		sending.value = false;
	}
};

// 开始倒计时
const startCountdown = () => {
	countdown.value = 60;
	if (countdownTimer) clearInterval(countdownTimer);
	countdownTimer = setInterval(() => {
		countdown.value--;
		if (countdown.value <= 0) {
			if (countdownTimer) clearInterval(countdownTimer);
		}
	}, 1000);
};

// 进入步骤3
const goToStep3 = () => {
	if (verificationCode.value.length !== 6) {
		codeError.value = '请输入6位验证码';
		return;
	}
	codeError.value = '';
	currentStep.value = 3;
};

// 重置密码
const resetPassword = async () => {
	if (!canSubmit.value) return;
	if (!resetToken.value) {
		uni.showToast({ title: '重置令牌无效，请重新获取验证码', icon: 'none' });
		currentStep.value = 1;
		return;
	}
	
	submitting.value = true;
	try {
		await authApi.resetPassword(resetToken.value, verificationCode.value, newPassword.value);
		currentStep.value = 4;
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : '重置失败';
		uni.showToast({ title: errorMessage, icon: 'none', duration: 2000 });
	} finally {
		submitting.value = false;
	}
};

// 返回登录
const goToLogin = () => {
	uni.redirectTo({
		url: '/pages/login/index'
	});
};

// 返回上一步
const handleBack = () => {
	if (currentStep.value > 1 && currentStep.value < 4) {
		currentStep.value--;
	} else {
		// 尝试返回，如果失败则跳转到登录页
		const pages = getCurrentPages();
		if (pages.length > 1) {
			uni.navigateBack();
		} else {
			uni.redirectTo({
				url: '/pages/login/index'
			});
		}
	}
};
</script>

<style lang="scss" scoped>
.forgot-page {
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

.forgot-content {
	position: relative;
	z-index: 10;
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	padding: 40rpx 48rpx;
	padding-top: calc(var(--status-bar-height) + 40rpx);
	box-sizing: border-box;
}

.back-btn {
	width: 80rpx;
	height: 80rpx;
	border-radius: 50%;
	background-color: #fff;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.1);
	margin-bottom: 40rpx;
	
	.iconfont {
		font-size: 40rpx;
		color: $text-primary;
	}
}

.step-indicator {
	display: flex;
	justify-content: center;
	gap: 16rpx;
	margin-bottom: 60rpx;
}

.step-dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	background-color: #e5e7eb;
	transition: all 0.3s;
	
	&.active {
		background-color: $primary;
	}
	
	&.current {
		width: 48rpx;
		border-radius: 8rpx;
	}
}

.step-content {
	flex: 1;
}

.step-header {
	margin-bottom: 60rpx;
}

.step-title {
	display: block;
	font-size: 48rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 16rpx;
}

.step-subtitle {
	display: block;
	font-size: 28rpx;
	color: $text-secondary;
}

.form-section {
	width: 100%;
}

.input-wrapper {
	position: relative;
	margin-bottom: 32rpx;
	
	&.input-error {
		.input-field {
			border-color: #ef4444;
		}
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

.resend-btn {
	position: absolute;
	right: 16rpx;
	top: 50%;
	transform: translateY(-50%);
	height: 64rpx;
	padding: 0 24rpx;
	font-size: 24rpx;
	color: $primary;
	background-color: transparent;
	border: 2rpx solid $primary;
	border-radius: 32rpx;
	
	&.disabled {
		color: $text-light;
		border-color: $border-color;
	}
}

.error-hint {
	display: block;
	font-size: 24rpx;
	color: #ef4444;
	margin-top: -16rpx;
	margin-bottom: 16rpx;
	padding-left: 16rpx;
}

.password-strength {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-bottom: 32rpx;
}

.strength-bars {
	display: flex;
	gap: 8rpx;
}

.strength-bar {
	width: 48rpx;
	height: 8rpx;
	border-radius: 4rpx;
	background-color: #e5e7eb;
	transition: background-color 0.3s;
}

.strength-text {
	font-size: 24rpx;
	
	&.weak { color: #ef4444; }
	&.medium { color: #f59e0b; }
	&.strong { color: #10b981; }
	&.very-strong { color: #3b82f6; }
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
	margin-top: 40rpx;
	
	&:active {
		transform: scale(0.98);
		opacity: 0.9;
	}
	
	&[disabled] {
		background-color: #9ca3af;
		box-shadow: none;
	}
}

.success-content {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding-top: 120rpx;
}

.success-icon-box {
	width: 160rpx;
	height: 160rpx;
	border-radius: 50%;
	background-color: #10b981;
	display: flex;
	align-items: center;
	justify-content: center;
	margin-bottom: 40rpx;
	
	.iconfont {
		font-size: 80rpx;
		color: #fff;
	}
}

.success-title {
	font-size: 40rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 16rpx;
}

.success-subtitle {
	font-size: 28rpx;
	color: $text-secondary;
	margin-bottom: 60rpx;
}
</style>
