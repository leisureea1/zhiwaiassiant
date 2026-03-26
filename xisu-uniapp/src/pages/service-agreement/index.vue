<template>
	<view class="agreement-page">
		<view class="nav-bar">
			<view class="back-btn" @tap="handleBack">
				<text class="iconfont icon-arrow_back"></text>
			</view>
			<text class="nav-title">用户服务协议</text>
			<view class="nav-placeholder"></view>
		</view>

		<scroll-view class="agreement-content" scroll-y>
			<view class="content-inner">
				<text class="update-date">更新日期：2026年2月1日</text>
				<text class="effective-date">生效日期：2026年2月1日</text>

				<view class="section">
					<text class="section-title">一、总则</text>
					<text class="section-text">1.1 知外助手（以下简称"本应用"）是提供教务信息查询等服务的工具类小程序。</text>
					<text class="section-text">1.2 用户在使用本应用前，应当仔细阅读本协议。请您务必审慎阅读、充分理解各条款内容。如您不同意本服务协议的任意内容，请您不要注册或使用本应用服务。</text>
					<text class="section-text">1.3 您在本应用进行注册、登录、使用等行为即视为您已阅读并同意受本协议的约束。</text>
				</view>

				<view class="section">
					<text class="section-title">二、服务内容</text>
					<text class="section-text">2.1 本应用为用户提供以下服务：</text>
					<text class="section-text indent">· 教务系统成绩查询；</text>
					<text class="section-text indent">· 考试安排查询；</text>
					<text class="section-text indent">· 课程信息查询；</text>
					<text class="section-text indent">· 评教服务；</text>
					<text class="section-text indent">· 校车时刻查询；</text>
					<text class="section-text indent">· AI 智能助手等。</text>
					<text class="section-text">2.2 本应用有权随时对服务内容进行调整，包括但不限于增加、减少或修改服务功能。</text>
				</view>

				<view class="section">
					<text class="section-title">三、用户账号</text>
					<text class="section-text">3.1 用户在注册时应提供真实、准确、完整的个人资料，并在资料发生变更时及时更新。</text>
					<text class="section-text">3.2 用户应妥善保管自己的账号和密码，因用户保管不善造成的损失由用户自行承担。</text>
					<text class="section-text">3.3 用户不得将账号转让、出售或以其他方式许可他人使用。</text>
				</view>

				<view class="section">
					<text class="section-title">四、用户行为规范</text>
					<text class="section-text">4.1 用户在使用本应用服务时，应遵守中华人民共和国相关法律法规。</text>
					<text class="section-text">4.2 用户不得利用本应用从事任何违法违规活动，包括但不限于：</text>
					<text class="section-text indent">· 非法获取或泄露他人个人信息；</text>
					<text class="section-text indent">· 干扰或破坏本应用及服务器的正常运行；</text>
					<text class="section-text indent">· 利用本应用进行任何可能损害第三方权益的行为。</text>
				</view>

				<view class="section">
					<text class="section-title">五、免责声明</text>
					<text class="section-text">5.1 本应用所展示的教务信息来源于学校教务系统，仅供参考，如有差异以学校官方信息为准。</text>
					<text class="section-text">5.2 因不可抗力、网络故障、系统维护等原因导致服务中断或数据延迟，本应用不承担相关责任。</text>
					<text class="section-text">5.3 用户因自身原因导致的任何损失，本应用不承担责任。</text>
				</view>

				<view class="section">
					<text class="section-title">六、知识产权</text>
					<text class="section-text">6.1 本应用的所有内容，包括但不限于文字、图片、代码、界面设计等，均受中华人民共和国知识产权法律的保护。</text>
					<text class="section-text">6.2 未经本应用事先书面同意，用户不得对本应用内容进行复制、修改、传播或以其他方式使用。</text>
				</view>

				<view class="section">
					<text class="section-title">七、协议的修改</text>
					<text class="section-text">7.1 本应用有权在必要时修改本协议条款。修改后的协议一旦公布即生效，代替原协议。</text>
					<text class="section-text">7.2 用户如不同意修改后的协议，可以主动停止使用本应用服务；如继续使用，则视为接受修改后的协议。</text>
				</view>

				<view class="section">
					<text class="section-title">八、其他</text>
					<text class="section-text">8.1 本协议的订立、执行和解释均适用中华人民共和国法律。</text>
					<text class="section-text">8.2 如本协议中的任何条款因任何原因被认定为无效或不可执行，其余条款仍然有效。</text>
				</view>
			</view>
		</scroll-view>
	</view>
</template>

<script setup lang="ts">
import { safeNavigateBack } from '@/utils/navigation';

const handleBack = () => {
	safeNavigateBack('/pages/login/index');
};
</script>

<style lang="scss" scoped>
.agreement-page {
	min-height: 100vh;
	background-color: #fff;
	display: flex;
	flex-direction: column;
}

.nav-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 32rpx;
	padding-top: calc(var(--status-bar-height) + 20rpx);
	height: calc(var(--status-bar-height) + 88rpx);
	background-color: #fff;
	border-bottom: 1rpx solid #f0f0f0;
	box-sizing: content-box;
}

.back-btn {
	width: 72rpx;
	height: 72rpx;
	display: flex;
	align-items: center;
	justify-content: center;

	.iconfont {
		font-size: 40rpx;
		color: #333;
	}
}

.nav-title {
	font-size: 34rpx;
	font-weight: 600;
	color: #333;
}

.nav-placeholder {
	width: 72rpx;
}

.agreement-content {
	flex: 1;
	height: 0;
}

.content-inner {
	padding: 40rpx 48rpx 80rpx;
}

.update-date,
.effective-date {
	display: block;
	font-size: 24rpx;
	color: #999;
	margin-bottom: 8rpx;
}

.effective-date {
	margin-bottom: 40rpx;
}

.section {
	margin-bottom: 40rpx;
}

.section-title {
	display: block;
	font-size: 32rpx;
	font-weight: 700;
	color: #333;
	margin-bottom: 20rpx;
}

.section-text {
	display: block;
	font-size: 28rpx;
	color: #666;
	line-height: 1.8;
	margin-bottom: 12rpx;

	&.indent {
		padding-left: 32rpx;
	}
}
</style>
