<template>
	<view class="ai-chat-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-content">
				<view class="back-btn" @tap="goBack">
					<text class="iconfont icon-arrow_back"></text>
				</view>
				<view class="header-title">
					<text class="iconfont icon-auto_awesome_mosaic"></text>
					<text class="title-text">Smart Assistant</text>
				</view>
				<view class="placeholder"></view>
			</view>
		</view>

		<!-- 消息列表 -->
		<scroll-view 
			class="messages-container" 
			scroll-y 
			:scroll-into-view="scrollToId"
			scroll-with-animation
		>
			<view 
				v-for="(msg, index) in messages" 
				:key="msg.id"
				:id="'msg-' + msg.id"
				class="message-wrapper"
				:class="msg.role"
			>
				<view class="message-bubble" :class="msg.role">
					<text class="message-text">{{ msg.text }}</text>
				</view>
			</view>

			<!-- 加载动画 -->
			<view v-if="isLoading" class="message-wrapper assistant">
				<view class="message-bubble assistant loading">
					<view class="loading-dots">
						<view class="dot"></view>
						<view class="dot"></view>
						<view class="dot"></view>
					</view>
				</view>
			</view>

			<view id="scroll-bottom"></view>
		</scroll-view>

		<!-- 输入区 -->
		<view class="input-area">
			<view class="input-wrapper">
				<input 
					class="input-field" 
					type="text" 
					v-model="inputText"
					placeholder="Ask anything..."
					placeholder-class="placeholder"
					@confirm="handleSend"
					:disabled="isLoading"
				/>
				<view 
					class="send-btn" 
					:class="{ active: inputText.trim() }"
					@tap="handleSend"
				>
					<text class="iconfont icon-arrow_forward"></text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { ChatMessage } from '@/types/index';
import { getUserInfo } from '@/services/apiService';

const userName = ref('用户');

onMounted(() => {
	const userInfo = getUserInfo<{ realName?: string; username?: string }>();
	if (userInfo) {
		userName.value = userInfo.realName || userInfo.username || '用户';
		messages.value[0].text = `Hi ${userName.value}! I'm your Zhiwai Smart Assistant. Ask me about your schedule, campus events, or study tips!`;
	}
});

const messages = ref<ChatMessage[]>([
	{
		id: '1',
		role: 'assistant',
		text: `Hi! I'm your Zhiwai Smart Assistant. Ask me about your schedule, campus events, or study tips!`,
		timestamp: new Date()
	}
]);

const inputText = ref('');
const isLoading = ref(false);
const scrollToId = ref('scroll-bottom');

const goBack = () => {
	safeNavigateBack();
};

const scrollToBottom = () => {
	nextTick(() => {
		scrollToId.value = '';
		setTimeout(() => {
			scrollToId.value = 'scroll-bottom';
		}, 50);
	});
};

const handleSend = async () => {
	const text = inputText.value.trim();
	if (!text || isLoading.value) return;

	// 添加用户消息
	const userMsg: ChatMessage = {
		id: Date.now().toString(),
		role: 'user',
		text: text,
		timestamp: new Date()
	};
	messages.value.push(userMsg);
	inputText.value = '';
	scrollToBottom();

	// 显示加载状态
	isLoading.value = true;

	// 模拟AI响应（实际项目中应调用API）
	setTimeout(() => {
		const responseText = generateMockResponse(text);
		const botMsg: ChatMessage = {
			id: (Date.now() + 1).toString(),
			role: 'assistant',
			text: responseText,
			timestamp: new Date()
		};
		messages.value.push(botMsg);
		isLoading.value = false;
		scrollToBottom();
	}, 1000 + Math.random() * 1000);
};

// 模拟AI响应
const generateMockResponse = (input: string): string => {
	const responses = [
		'好的，让我帮你查一下相关信息。',
		'这是一个很好的问题！根据校园信息，我可以告诉你...',
		'我理解你的问题。关于这个，我建议你可以...',
		'感谢你的提问！以下是我找到的相关信息：',
		'好的，我来为你解答这个问题。'
	];
	
	if (input.includes('课') || input.includes('schedule')) {
		return '你今天有3节课：上午的实践俄语III（8:00-9:50，JB107），下午的国际法学（14:00-15:50，JB107），以及晚上的毛泽东思想（19:10-21:00，JA410）。';
	}
	
	if (input.includes('考试') || input.includes('exam')) {
		return '你最近的考试是1月10日的概率论与数理统计，在教二JA101，时间是9:00-11:00。还有3天时间准备！';
	}
	
	if (input.includes('成绩') || input.includes('grade')) {
		return '你本学期的GPA是3.86，平均分91.5。最高分是C语言程序设计（98分），继续保持！';
	}
	
	return responses[Math.floor(Math.random() * responses.length)];
};
</script>

<style lang="scss" scoped>
.ai-chat-page {
	display: flex;
	flex-direction: column;
	height: 100vh;
	background-color: $bg-light;
}

.header {
	position: sticky;
	top: 0;
	z-index: 50;
	background: rgba(248, 250, 252, 0.9);
	backdrop-filter: blur(20rpx);
	border-bottom: 2rpx solid $border-color;
	padding: 24rpx 32rpx;
	padding-top: calc(var(--status-bar-height) + 48rpx);
}

.header-content {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.back-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	
	.icon {
		font-size: 40rpx;
		color: $text-primary;
	}
	
	&:active {
		background-color: rgba(0,0,0,0.05);
	}
}

.header-title {
	display: flex;
	align-items: center;
	gap: 12rpx;
	
	.icon {
		font-size: 36rpx;
		color: $primary;
	}
	
	.title-text {
		font-size: 40rpx;
		font-weight: 700;
		color: $text-primary;
	}
}

.placeholder {
	width: 64rpx;
}

.messages-container {
	flex: 1;
	padding: 24rpx 32rpx;
	box-sizing: border-box;
}

.message-wrapper {
	display: flex;
	margin-bottom: 24rpx;
	
	&.user {
		justify-content: flex-end;
	}
	
	&.assistant {
		justify-content: flex-start;
	}
}

.message-bubble {
	max-width: 80%;
	padding: 24rpx 32rpx;
	border-radius: 32rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.05);
	
	&.user {
		background-color: $primary;
		border-bottom-right-radius: 8rpx;
		
		.message-text {
			color: #fff;
		}
	}
	
	&.assistant {
		background-color: #fff;
		border: 2rpx solid $border-color;
		border-bottom-left-radius: 8rpx;
		
		.message-text {
			color: $text-primary;
		}
	}
	
	&.loading {
		padding: 24rpx 32rpx;
	}
}

.message-text {
	font-size: 28rpx;
	line-height: 1.6;
}

.loading-dots {
	display: flex;
	gap: 12rpx;
}

.dot {
	width: 16rpx;
	height: 16rpx;
	background-color: $primary;
	border-radius: 50%;
	animation: bounce 1.4s infinite ease-in-out both;
	
	&:nth-child(1) {
		animation-delay: 0ms;
	}
	
	&:nth-child(2) {
		animation-delay: 160ms;
	}
	
	&:nth-child(3) {
		animation-delay: 320ms;
	}
}

@keyframes bounce {
	0%, 80%, 100% {
		transform: scale(0.6);
		opacity: 0.6;
	}
	40% {
		transform: scale(1);
		opacity: 1;
	}
}

.input-area {
	position: sticky;
	bottom: 0;
	background-color: $bg-light;
	border-top: 2rpx solid $border-color;
	padding: 24rpx 32rpx;
	padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
}

.input-wrapper {
	display: flex;
	align-items: center;
	gap: 16rpx;
	background-color: #fff;
	border-radius: 48rpx;
	border: 2rpx solid $border-color;
	padding: 8rpx 16rpx 8rpx 32rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
}

.input-field {
	flex: 1;
	height: 80rpx;
	font-size: 28rpx;
	color: $text-primary;
	background: transparent;
}

.placeholder {
	color: $text-light;
}

.send-btn {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	background-color: #f3f4f6;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.2s;
	
	.icon {
		font-size: 32rpx;
		color: $text-light;
	}
	
	&.active {
		background-color: $primary;
		
		.icon {
			color: #fff;
		}
	}
	
	&:active.active {
		transform: scale(0.95);
	}
}
</style>
