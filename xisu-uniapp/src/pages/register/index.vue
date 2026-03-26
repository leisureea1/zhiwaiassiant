<template>
	<view class="register-page">
		<!-- 点阵背景 -->
		<view class="dot-pattern"></view>

		<view class="register-content">
			<!-- 返回按钮 -->
			<view class="back-btn" @tap="handleBack">
				<text class="iconfont icon-arrow_back"></text>
			</view>

			<!-- 步骤指示器 -->
			<view class="step-indicator">
				<view 
					v-for="i in 4" 
					:key="i" 
					class="step-dot"
					:class="{ active: currentStep >= i, current: currentStep === i }"
				></view>
			</view>

			<!-- 步骤1: 用户名和密码 -->
			<view v-if="currentStep === 1" class="step-content">
				<view class="step-header">
					<text class="step-title">创建账号</text>
					<text class="step-subtitle">设置您的用户名和密码</text>
				</view>

				<view class="form-section">
					<view class="input-wrapper" :class="{ 'input-error': usernameError }">
						<view class="input-icon">
							<text class="iconfont icon-account_circle"></text>
						</view>
						<input 
							class="input-field" 
							type="text" 
							v-model="username"
							placeholder="用户名 (字母数字组合)" 
							placeholder-class="placeholder"
							@input="validateUsername"
						/>
					</view>
					<text v-if="usernameError" class="error-hint">{{ usernameError }}</text>

					<view class="input-wrapper" :class="{ 'input-error': passwordError }">
						<view class="input-icon">
							<text class="iconfont icon-flight_class"></text>
						</view>
						<input 
							class="input-field" 
							type="text"
							:password="!showPassword"
							v-model="password"
							placeholder="设置密码 (至少8位，含大小写字母和数字)" 
							placeholder-class="placeholder"
							@input="validatePassword"
						/>
						<view class="toggle-password" @tap="showPassword = !showPassword">
							<text class="iconfont" :class="showPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
					<text v-if="passwordError" class="error-hint">{{ passwordError }}</text>
					
					<!-- 密码强度指示器 -->
					<view v-if="password && !passwordError" class="password-strength">
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
							placeholder="确认密码" 
							placeholder-class="placeholder"
							@input="validateConfirmPassword"
						/>
						<view class="toggle-password" @tap="showConfirmPassword = !showConfirmPassword">
							<text class="iconfont" :class="showConfirmPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>
					<text v-if="confirmPasswordError" class="error-hint">{{ confirmPasswordError }}</text>

					<!-- 隐私协议勾选 -->
					<view class="agreement-section">
						<view class="agreement-row" @tap="agreedPrivacy = !agreedPrivacy">
							<view class="checkbox">
								<text v-if="agreedPrivacy" class="iconfont icon-check_circle" style="font-size: 36rpx; color: #3B82F6;"></text>
								<view v-else class="checkbox-empty"></view>
							</view>
							<text class="agreement-text">我已阅读并同意</text>
						</view>
						<view class="agreement-links">
							<text class="agreement-link" @tap.stop="openServiceAgreement">《用户服务协议》</text>
							<text class="agreement-text">和</text>
							<text class="agreement-link" @tap.stop="openPrivacyPolicy">《隐私政策》</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 步骤2: 邮箱 -->
			<view v-if="currentStep === 2" class="step-content">
				<view class="step-header">
					<text class="step-title">验证邮箱</text>
					<text class="step-subtitle">用于找回密码和接收通知</text>
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
							placeholder="邮箱地址" 
							placeholder-class="placeholder"
							@blur="validateEmail"
							:disabled="emailVerified"
						/>
						<view v-if="emailVerified" class="verified-badge">
							<text class="iconfont icon-check_circle" style="color: #10b981;"></text>
						</view>
					</view>
					<text v-if="emailError" class="error-hint">{{ emailError }}</text>

					<!-- 验证码输入区域 -->
					<view v-if="!emailVerified" class="verification-section">
						<view class="input-wrapper code-input-wrapper">
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
								:disabled="!codeSent"
							/>
							<button 
								class="send-code-btn"
								:class="{ disabled: countdown > 0 || !email || emailError }"
								:disabled="countdown > 0 || !email || !!emailError"
								@tap="sendVerificationCode"
							>
								{{ countdown > 0 ? `${countdown}s` : (codeSent ? '重新发送' : '发送验证码') }}
							</button>
						</view>
						<text v-if="codeError" class="error-hint">{{ codeError }}</text>

						<button 
							v-if="codeSent && verificationCode.length === 6"
							class="verify-btn"
							@tap="verifyCode"
							:disabled="verifying"
						>
							{{ verifying ? '验证中...' : '验证' }}
						</button>
					</view>

					<view v-if="emailVerified" class="success-box">
						<text class="iconfont icon-check_circle success-icon"></text>
						<text class="success-text">邮箱验证成功！</text>
					</view>

					<view v-else class="tips-box">
						<text class="iconfont icon-info tips-icon"></text>
						<text class="tips-text">验证码将发送到您的邮箱，10分钟内有效</text>
					</view>
				</view>
			</view>

			<!-- 步骤3: 学号和掌上西外密码 -->
			<view v-if="currentStep === 3" class="step-content">
				<view class="step-header">
					<text class="step-title">绑定学号</text>
					<text class="step-subtitle">关联您的西外账号</text>
				</view>

				<view class="form-section">
					<view class="input-wrapper">
						<view class="input-icon">
							<text class="iconfont icon-assignment_ind"></text>
						</view>
						<input 
							class="input-field" 
							type="text" 
							v-model="studentId"
							placeholder="学号 / Student ID" 
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
							:password="!showXiwaiPassword"
							v-model="xiwaiPassword"
							placeholder="掌上西外系统密码" 
							placeholder-class="placeholder"
						/>
						<view class="toggle-password" @tap="showXiwaiPassword = !showXiwaiPassword">
							<text class="iconfont" :class="showXiwaiPassword ? 'icon-visibility' : 'icon-visibility_off'"></text>
						</view>
					</view>

					<view class="tips-box">
						<text class="iconfont icon-info tips-icon"></text>
						<text class="tips-text">密码仅用于验证身份，不会被保存</text>
					</view>
				</view>
			</view>

			<!-- 步骤4: 头像上传 -->
			<view v-if="currentStep === 4" class="step-content">
				<view class="step-header">
					<text class="step-title">设置头像</text>
					<text class="step-subtitle">让大家认识你</text>
				</view>

				<view class="avatar-section">
					<!-- 微信小程序使用 chooseAvatar -->
					<button class="avatar-upload-btn" open-type="chooseAvatar" @chooseavatar="onChooseAvatar">
						<image v-if="avatarUrl" :src="avatarUrl" class="avatar-preview" mode="aspectFill" />
						<view v-else class="avatar-placeholder">
							<text class="iconfont icon-add_photo"></text>
							<text class="upload-text">点击获取微信头像</text>
						</view>
					</button>

					<view class="avatar-tips">
						<text class="tips-text">使用微信头像，一键获取</text>
					</view>

					<view class="skip-btn" @tap="skipAvatar">
						<text class="skip-text">暂时跳过，使用默认头像</text>
					</view>
				</view>
			</view>

			<!-- 底部按钮 -->
			<view class="button-section">
				<button 
					v-if="currentStep < 4" 
					class="btn-primary" 
					@tap="nextStep"
				>
					下一步
				</button>
				<button 
					v-else 
					class="btn-primary" 
					@tap="handleRegister"
					:disabled="isRegistering"
				>
					{{ isRegistering ? '注册中...' : '完成注册' }}
				</button>

				<view v-if="currentStep === 1" class="login-hint">
					<text class="hint-text">已有账号？</text>
					<text class="login-link" @tap="goToLogin">立即登录</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { authApi, userApi, saveTokens, saveUserInfo } from '@/services/apiService';
import { debugLog, debugWarn } from '@/utils/debug';
import { calcPasswordStrength, strengthLabels, strengthClasses, strengthColors, validatePasswordFormat, validatePasswordMatch } from '@/utils/password';

// 当前步骤
const currentStep = ref(1);
const isRegistering = ref(false);
const agreedPrivacy = ref(false);

const openServiceAgreement = () => {
	uni.navigateTo({ url: '/pages/service-agreement/index' });
};

const openPrivacyPolicy = () => {
	uni.navigateTo({ url: '/pages/privacy-policy/index' });
};

// 步骤1: 用户名密码
const username = ref('');
const password = ref('');
const confirmPassword = ref('');
const showPassword = ref(false);
const showConfirmPassword = ref(false);
const usernameError = ref('');
const passwordError = ref('');
const confirmPasswordError = ref('');

// 密码强度计算
const passwordStrength = computed(() => calcPasswordStrength(password.value));

const strengthText = computed(() => strengthLabels[passwordStrength.value]);

const strengthClass = computed(() => strengthClasses[passwordStrength.value]);


const strengthColor = computed(() => strengthColors[passwordStrength.value]);
});

// 步骤2: 邮箱
const email = ref('');
const emailError = ref('');
const verificationCode = ref('');
const codeError = ref('');
const codeSent = ref(false);
const countdown = ref(0);
const emailVerified = ref(false);
const emailToken = ref('');
const verifying = ref(false);
let countdownTimer: ReturnType<typeof setInterval> | null = null;

onUnmounted(() => {
	if (countdownTimer) {
		clearInterval(countdownTimer);
		countdownTimer = null;
	}
});

// 发送验证码
const sendVerificationCode = async () => {
	if (!email.value || emailError.value || countdown.value > 0) return;
	
	validateEmail();
	if (emailError.value) return;
	
	try {
		await authApi.sendCode({ email: email.value });
		
		codeSent.value = true;
		codeError.value = '';
		countdown.value = 60;
		
		// 开始倒计时
		countdownTimer = setInterval(() => {
			countdown.value--;
			if (countdown.value <= 0) {
				if (countdownTimer) clearInterval(countdownTimer);
			}
		}, 1000);
		
		uni.showToast({ title: '验证码已发送', icon: 'success' });
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : '发送失败，请稍后再试';
		codeError.value = errorMessage;
		uni.showToast({ title: errorMessage, icon: 'none' });
	}
};

// 验证验证码
const verifyCode = async () => {
	if (verificationCode.value.length !== 6) {
		codeError.value = '请输入6位验证码';
		return;
	}
	
	verifying.value = true;
	try {
		const res = await authApi.verifyCode({
			email: email.value,
			code: verificationCode.value,
		});
		
		debugLog('[verifyCode] Response received:', JSON.stringify(res));
		debugLog('[verifyCode] res.verified =', res.verified, 'type:', typeof res.verified);
		
		if (res.verified) {
			emailVerified.value = true;
			emailToken.value = res.emailToken || '';
			codeError.value = '';
			if (countdownTimer) clearInterval(countdownTimer);
			uni.showToast({ title: '邮箱验证成功', icon: 'success' });
		} else {
			codeError.value = '验证码错误';
			uni.showToast({ title: '验证失败', icon: 'none' });
		}
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : '网络错误，请稍后再试';
		codeError.value = errorMessage;
		uni.showToast({ title: errorMessage, icon: 'none' });
	} finally {
		verifying.value = false;
	}
};

// 步骤3: 学号
const studentId = ref('');
const xiwaiPassword = ref('');
const showXiwaiPassword = ref(false);

// 步骤4: 头像
const avatarUrl = ref('');

// 用户名验证
const validateUsername = () => {
	const value = username.value;
	if (!value) {
		usernameError.value = '';
		return;
	}
	const regex = /^[a-zA-Z0-9]+$/;
	if (!regex.test(value)) {
		usernameError.value = '用户名只能包含字母和数字';
	} else if (value.length < 3) {
		usernameError.value = '用户名至少3个字符';
	} else if (value.length > 20) {
		usernameError.value = '用户名最多20个字符';
	} else {
		usernameError.value = '';
	}
};

// 密码验证
const validatePassword = () => {
	const value = password.value;
	if (!value) {
		passwordError.value = '';
		return;
	}
	if (value.length > 20) {
		passwordError.value = '密码最多20位';
	} else {
		passwordError.value = validatePasswordFormat(value);
	}
	if (confirmPassword.value) {
		validateConfirmPassword();
	}
};

// 确认密码验证
const validateConfirmPassword = () => {
	if (!confirmPassword.value) {
		confirmPasswordError.value = '';
		return;
	}
	confirmPasswordError.value = validatePasswordMatch(password.value, confirmPassword.value);
};
	}
};

// 邮箱验证
const validateEmail = () => {
	const value = email.value;
	if (!value) {
		emailError.value = '';
		return;
	}
	const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
	if (!regex.test(value)) {
		emailError.value = '请输入有效的邮箱地址';
	} else {
		emailError.value = '';
	}
};

// 返回上一步或退出
const handleBack = () => {
	if (currentStep.value > 1) {
		currentStep.value--;
	} else {
		safeNavigateBack('/pages/login/index');
	}
};

// 下一步
const nextStep = () => {
	switch (currentStep.value) {
		case 1:
			validateUsername();
			validatePassword();
			validateConfirmPassword();
			if (!username.value) {
				uni.showToast({ title: '请输入用户名', icon: 'none' });
				return;
			}
			if (!password.value) {
				uni.showToast({ title: '请设置密码', icon: 'none' });
				return;
			}
			if (!confirmPassword.value) {
				uni.showToast({ title: '请确认密码', icon: 'none' });
				return;
			}
			if (usernameError.value || passwordError.value || confirmPasswordError.value) {
				return;
			}
			if (!agreedPrivacy.value) {
				uni.showToast({ title: '请先阅读并同意用户服务协议和隐私政策', icon: 'none', duration: 2000 });
				return;
			}
			break;
		case 2:
			validateEmail();
			if (!email.value) {
				uni.showToast({ title: '请输入邮箱', icon: 'none' });
				return;
			}
			if (emailError.value) {
				return;
			}
			if (!emailVerified.value) {
				uni.showToast({ title: '请先完成邮箱验证', icon: 'none' });
				return;
			}
			break;
		case 3:
			if (!studentId.value) {
				uni.showToast({ title: '请输入学号', icon: 'none' });
				return;
			}
			if (!xiwaiPassword.value) {
				uni.showToast({ title: '请输入掌上西外密码', icon: 'none' });
				return;
			}
			break;
	}
	currentStep.value++;
};

// 微信获取头像回调
const onChooseAvatar = (e: any) => {
	avatarUrl.value = e.detail.avatarUrl;
};

// 跳过头像
const skipAvatar = () => {
	avatarUrl.value = '';
	handleRegister();
};

// 去登录
const goToLogin = () => {
	safeNavigateBack('/pages/login/index');
};

// 完成注册
const handleRegister = async () => {
	if (isRegistering.value) return;
	
	isRegistering.value = true;
	uni.showLoading({ title: '注册中...' });
	
	try {
		const res = await authApi.register({
			username: username.value,
			password: password.value,
			email: email.value,
			studentId: studentId.value,
			xiwaiPassword: xiwaiPassword.value,
			emailToken: emailToken.value,
			avatar:
				avatarUrl.value && !avatarUrl.value.startsWith('wxfile://')
					? avatarUrl.value
					: undefined,
		});
		
		// 保存登录凭证和用户信息
		saveTokens({
			accessToken: res.accessToken,
			refreshToken: res.refreshToken,
		});

		// 对微信临时头像执行真实上传并覆盖用户头像地址
		if (avatarUrl.value && avatarUrl.value.startsWith('wxfile://') && res.user?.id) {
			try {
				uni.showLoading({ title: '上传头像中...' });
				const uploaded = await userApi.uploadAvatarFile(res.user.id, avatarUrl.value);
				if (uploaded?.avatar) {
					res.user.avatar = uploaded.avatar;
				}
			} catch (uploadError) {
				debugWarn('[Register] avatar upload failed:', uploadError);
				uni.showToast({ title: '头像上传失败，已使用默认头像', icon: 'none' });
			}
		}

		saveUserInfo(res.user);
		
		uni.hideLoading();
		uni.showToast({ 
			title: '注册成功', 
			icon: 'success',
			duration: 1500
		});
		
		setTimeout(() => {
			uni.switchTab({ url: '/pages/home/index' });
		}, 1500);
	} catch (error) {
		uni.hideLoading();
		const errorMessage = error instanceof Error ? error.message : '注册失败，请稍后再试';
		uni.showToast({ title: errorMessage, icon: 'none', duration: 2000 });
	} finally {
		isRegistering.value = false;
	}
};
</script>

<style lang="scss" scoped>
.register-page {
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

.register-content {
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
	position: absolute;
	top: calc(var(--status-bar-height) + 20rpx);
	left: 32rpx;
	width: 72rpx;
	height: 72rpx;
	border-radius: 20rpx;
	background-color: #fff;
	box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.08);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 20;
	
	.iconfont {
		font-size: 40rpx;
		color: $text-secondary;
	}
}

// 步骤指示器
.step-indicator {
	display: flex;
	justify-content: center;
	gap: 16rpx;
	margin-top: 80rpx;
	margin-bottom: 48rpx;
}

.step-dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	background-color: #e5e7eb;
	transition: all 0.3s ease;
	
	&.active {
		background-color: $primary;
	}
	
	&.current {
		width: 48rpx;
		border-radius: 8rpx;
	}
}

// 步骤内容
.step-content {
	flex: 1;
}

.step-header {
	text-align: center;
	margin-bottom: 60rpx;
}

.step-title {
	display: block;
	font-size: 48rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 12rpx;
}

.step-subtitle {
	display: block;
	font-size: 28rpx;
	color: $text-secondary;
}

// 表单
.form-section {
	width: 100%;
}

.input-wrapper {
	position: relative;
	margin-bottom: 24rpx;
	
	&.input-error .input-field {
		border-color: #ef4444;
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

.error-hint {
	display: block;
	font-size: 24rpx;
	color: #ef4444;
	margin-top: -16rpx;
	margin-bottom: 16rpx;
	padding-left: 16rpx;
}

// 密码强度指示器
.password-strength {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-top: -8rpx;
	margin-bottom: 16rpx;
	padding-left: 16rpx;
}

.strength-bars {
	display: flex;
	gap: 8rpx;
}

.strength-bar {
	width: 40rpx;
	height: 8rpx;
	border-radius: 4rpx;
	background-color: #e5e7eb;
	transition: all 0.3s ease;
	
	&.active {
		background-color: #ef4444;
	}
}

.strength-text {
	font-size: 22rpx;
	
	&.weak {
		color: #ef4444;
	}
	
	&.medium {
		color: #f59e0b;
	}
	
	&.strong {
		color: #10b981;
	}
	
	&.very-strong {
		color: #059669;
	}
}

.tips-box {
	display: flex;
	align-items: center;
	padding: 24rpx;
	background-color: rgba(59, 130, 246, 0.05);
	border-radius: 16rpx;
	margin-top: 24rpx;
	
	.tips-icon {
		font-size: 32rpx;
		color: $primary;
		margin-right: 16rpx;
	}
	
	.tips-text {
		font-size: 24rpx;
		color: $text-secondary;
		line-height: 1.4;
	}
}

// 头像上传
.avatar-section {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding-top: 40rpx;
}

.avatar-upload-btn {
	width: 240rpx;
	height: 240rpx;
	border-radius: 50%;
	background-color: #fff;
	border: 4rpx dashed $border-color;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	padding: 0;
	margin: 0;
	line-height: 1;
	
	&::after {
		border: none;
	}
	
	&:active {
		border-color: $primary;
	}
}

.avatar-preview {
	width: 100%;
	height: 100%;
}

.avatar-placeholder {
	display: flex;
	flex-direction: column;
	align-items: center;
	
	.iconfont {
		font-size: 64rpx;
		color: $text-light;
		margin-bottom: 12rpx;
	}
	
	.upload-text {
		font-size: 24rpx;
		color: $text-light;
	}
}

.avatar-tips {
	margin-top: 32rpx;
	
	.tips-text {
		font-size: 24rpx;
		color: $text-light;
	}
}

.skip-btn {
	margin-top: 48rpx;
	padding: 16rpx 32rpx;
	
	.skip-text {
		font-size: 28rpx;
		color: $text-secondary;
		text-decoration: underline;
	}
}

// 底部按钮
.button-section {
	padding-top: 40rpx;
	padding-bottom: 40rpx;
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

.login-hint {
	display: flex;
	justify-content: center;
	align-items: center;
	margin-top: 32rpx;
}

.hint-text {
	font-size: 28rpx;
	color: $text-secondary;
}

.login-link {
	font-size: 28rpx;
	color: $primary;
	font-weight: 600;
	margin-left: 8rpx;
}

// 验证码相关样式
.verification-section {
	margin-top: 24rpx;
}

.code-input-wrapper {
	position: relative;
	
	.input-field {
		padding-right: 200rpx;
	}
}

.send-code-btn {
	position: absolute;
	right: 16rpx;
	top: 50%;
	transform: translateY(-50%);
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: white;
	font-size: 24rpx;
	padding: 12rpx 24rpx;
	border-radius: 20rpx;
	border: none;
	line-height: 1.2;
	
	&.disabled {
		background: #d1d5db;
		color: #9ca3af;
	}
}

.verify-btn {
	width: 100%;
	height: 88rpx;
	background: linear-gradient(135deg, #10b981 0%, #059669 100%);
	color: white;
	font-size: 32rpx;
	font-weight: 600;
	border-radius: 44rpx;
	margin-top: 32rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border: none;
	
	&:active {
		opacity: 0.9;
	}
	
	&[disabled] {
		background: #d1d5db;
		color: #9ca3af;
	}
}

.verified-badge {
	position: absolute;
	right: 24rpx;
	top: 50%;
	transform: translateY(-50%);
	font-size: 40rpx;
}

.success-box {
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 32rpx;
	background: rgba(16, 185, 129, 0.1);
	border-radius: 16rpx;
	margin-top: 24rpx;
}

.success-icon {
	font-size: 48rpx;
	color: #10b981;
	margin-right: 16rpx;
}

.success-text {
	font-size: 32rpx;
	color: #10b981;
	font-weight: 600;
}

// 协议勾选区域
.agreement-section {
	margin-top: 32rpx;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
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
	padding-left: 44rpx;
}

.agreement-link {
	font-size: 24rpx;
	color: $primary;
}
</style>
