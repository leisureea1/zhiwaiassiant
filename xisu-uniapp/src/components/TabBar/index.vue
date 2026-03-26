<template>
	<view class="tab-bar" :style="{ paddingBottom: safeAreaBottom + 'px' }">
		<view 
			class="tab-item" 
			v-for="(item, index) in tabList" 
			:key="index"
			@click="switchTab(item.pagePath)"
		>
			<view class="icon-wrapper">
				<text 
					class="tab-icon"
					:style="{ color: current === index ? selectedColor : color }"
				>{{ item.icon }}</text>
			</view>
			<text 
				class="tab-text"
				:style="{ color: current === index ? selectedColor : color }"
			>{{ item.text }}</text>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const props = defineProps<{
	current: number
}>()

const safeAreaBottom = ref(0)
const color = '#999999'
const selectedColor = '#3B82F6'

// 使用 iconfont 的 Unicode 编码
const tabList = [
	{
		pagePath: '/pages/home/index',
		text: '首页',
		icon: '\ue60e'  // icon-home
	},
	{
		pagePath: '/pages/apps/index',
		text: '应用',
		icon: '\ue604'  // icon-apps
	},
	{
		pagePath: '/pages/profile/index',
		text: '我的',
		icon: '\ue605'  // icon-assignment_ind (用户)
	}
]

onMounted(() => {
	// 获取安全区域高度
	const systemInfo = uni.getSystemInfoSync()
	safeAreaBottom.value = systemInfo.safeAreaInsets?.bottom || 0
})

const switchTab = (path: string) => {
	uni.switchTab({ url: path })
}
</script>

<style lang="scss" scoped>
@import '@/static/iconfont/iconfont.css';

.tab-bar {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	height: 120rpx;
	background-color: #ffffff;
	display: flex;
	align-items: stretch;
	justify-content: space-around;
	box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
	z-index: 999;
}

.tab-item {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	flex: 1;
	height: 120rpx;
	transition: all 0.2s ease;
	
	&:active {
		transform: scale(0.95);
	}
}

.icon-wrapper {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 52rpx;
	margin-bottom: 4rpx;
}

.tab-icon {
	font-family: "iconfont" !important;
	font-size: 48rpx;
	font-style: normal;
	line-height: 1;
	transition: color 0.2s ease;
}

.tab-text {
	font-size: 22rpx;
	line-height: 1;
	transition: color 0.2s ease;
}
</style>
