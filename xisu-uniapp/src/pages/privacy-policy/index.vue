<template>
	<view class="policy-page">
		<view class="nav-bar">
			<view class="back-btn" @tap="handleBack">
				<text class="iconfont icon-arrow_back"></text>
			</view>
			<text class="nav-title">隐私政策</text>
			<view class="nav-placeholder"></view>
		</view>

		<scroll-view class="policy-content" scroll-y>
			<view class="content-inner">
				<text class="update-date">更新日期：2026年2月1日</text>
				<text class="effective-date">生效日期：2026年2月1日</text>

				<view class="section">
					<text class="section-title">引言</text>
					<text class="section-text">知外助手（以下简称"我们"）非常重视用户的隐私和个人信息保护。您在使用我们的服务时，我们可能会收集和使用您的相关信息。我们希望通过本《隐私政策》向您说明，在使用我们的服务时，我们如何收集、使用、储存和分享这些信息，以及我们为您提供的访问、更新、控制和保护这些信息的方式。</text>
					<text class="section-text">本《隐私政策》与您所使用的知外助手服务息息相关，希望您仔细阅读。您使用或继续使用我们的服务，即意味着同意我们按照本《隐私政策》收集、使用、储存和分享您的相关信息。</text>
				</view>

				<view class="section">
					<text class="section-title">一、我们收集的信息</text>
					<text class="section-text">我们提供服务时，可能会收集、储存和使用下列与您有关的信息：</text>
					<text class="section-text">1. 您提供的信息</text>
					<text class="section-text indent">· 您在注册账户时填写的信息，例如您的用户名、学号、邮箱地址；</text>
					<text class="section-text indent">· 您通过我们的服务向其他方提供的共享信息，以及您使用我们的服务时所储存的信息。</text>
					<text class="section-text">2. 我们获取的信息</text>
					<text class="section-text indent">· 日志信息：当您使用我们的服务时，我们可能会自动收集相关信息并存储为服务日志信息，包括设备信息、软件信息、IP地址等；</text>
					<text class="section-text indent">· 位置信息：仅在您主动授权时用于天气服务功能，不会后台持续获取。</text>
				</view>

				<view class="section">
					<text class="section-title">二、信息的使用</text>
					<text class="section-text">我们可能将收集到的信息用于以下用途：</text>
					<text class="section-text indent">· 向您提供服务（教务查询、成绩查询、考试安排等）；</text>
					<text class="section-text indent">· 满足您的个性化需求，如语言设置、位置设置、个性化帮助等；</text>
					<text class="section-text indent">· 在我们提供服务时用于身份验证、安全防范、诈骗监测、存档和备份；</text>
					<text class="section-text indent">· 帮助我们设计新服务，改善我们现有服务。</text>
				</view>

				<view class="section">
					<text class="section-title">三、信息的存储</text>
					<text class="section-text">我们会采取合理可行的措施，尽力避免收集无关的个人信息。我们只会在达成本政策所述目的所需的期限内保留您的个人信息，除非需要延长保留期或受到法律的允许。</text>
					<text class="section-text">我们在中华人民共和国境内收集和产生的个人信息，将存储在中华人民共和国境内。</text>
				</view>

				<view class="section">
					<text class="section-title">四、信息安全</text>
					<text class="section-text">我们使用各种安全技术和程序来防止信息的丢失、不当使用、未经授权阅览或披露。例如，我们会使用加密技术（如SSL）、匿名化处理等手段来保护您的个人信息。</text>
					<text class="section-text">尽管已经采取了合理有效的措施，并已经遵守了相关法律规定要求的标准，但请您理解，由于技术的限制以及可能存在的各种恶意手段，在互联网行业，即便竭尽所能加强安全措施，也不可能始终保证信息百分之百的安全。</text>
				</view>

				<view class="section">
					<text class="section-title">五、您的权利</text>
					<text class="section-text">您可以随时：</text>
					<text class="section-text indent">· 访问和修改您的个人信息；</text>
					<text class="section-text indent">· 删除您的账户和相关个人信息；</text>
					<text class="section-text indent">· 撤回您之前授予的同意。</text>
				</view>

				<view class="section">
					<text class="section-title">六、未成年人保护</text>
					<text class="section-text">我们鼓励父母或监护人指导未满十八岁的未成年人使用我们的服务。我们建议未成年人鼓励他们的父母或监护人阅读本《隐私政策》，并建议未成年人在提交个人信息之前寻求父母或监护人的同意和指导。</text>
				</view>

				<view class="section">
					<text class="section-title">七、隐私政策的变更</text>
					<text class="section-text">我们可能适时修订本《隐私政策》的条款，该等修订构成本《隐私政策》的一部分。如该等修订造成您在本《隐私政策》下权利的实质减少，我们将在修订生效前通过在主页上显著位置提示或向您发送通知等方式来通知您。</text>
				</view>

				<view class="section">
					<text class="section-title">八、联系我们</text>
					<text class="section-text">如您对本隐私政策或相关事宜有任何问题、意见或建议，请通过应用内反馈功能或发送邮件与我们联系。</text>
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
.policy-page {
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

.policy-content {
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
