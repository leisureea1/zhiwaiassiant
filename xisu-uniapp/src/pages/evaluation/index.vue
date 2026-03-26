<template>
	<view class="evaluation-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-content">
				<view class="back-btn" @tap="goBack">
					<text class="iconfont icon-arrow_back"></text>
				</view>
				<text class="page-title">评教系统</text>
				<view class="refresh-btn" @tap="handleRefresh">
					<text class="iconfont icon-arrow_forward"></text>
				</view>
			</view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- 标题区 -->
			<view class="section-header">
				<text class="section-title">待评教课程</text>
				<view class="pending-badge">
					<text class="badge-text">{{ evaluations.length }} 门待办</text>
				</view>
			</view>

			<!-- 加载状态 -->
			<view v-if="isLoading" class="loading-container">
				<text class="loading-text">加载中...</text>
			</view>

			<!-- 空状态 -->
			<view v-else-if="evaluations.length === 0" class="empty-container">
				<text class="iconfont icon-check_circle empty-icon"></text>
				<text class="empty-text">{{ emptyMessage }}</text>
			</view>

			<!-- 课程列表 -->
			<view v-else class="courses-list">
				<view 
					v-for="(item, index) in evaluations" 
					:key="item.lesson_id"
					class="course-card"
				>
					<view class="card-header">
						<view class="card-left">
							<view class="pending-tag">
								<text class="tag-text">待评教</text>
							</view>
							<text class="course-title">{{ item.course_name || '未知课程' }}</text>
							<text class="course-dept">{{ item.course_code }} · {{ item.course_type || '课程' }}</text>
						</view>
						<view class="course-icon" :class="getColorClass(index)">
							<text class="iconfont icon-menu_book"></text>
						</view>
					</view>
					
					<view class="card-divider"></view>
					
					<view class="card-footer">
						<view class="teacher-icon">
							<text class="iconfont icon-person"></text>
						</view>
						<view class="prof-info">
							<text class="prof-name">{{ item.teacher_name || '未知教师' }}</text>
							<text class="prof-role">主讲教师</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 加载完成提示 -->
			<view v-if="!isLoading && evaluations.length > 0" class="end-tip">
				<text class="tip-text">已加载全部待评教课程</text>
			</view>
		</scroll-view>

		<!-- 底部按钮 -->
		<view v-if="evaluations.length > 0" class="bottom-action">
			<button class="action-btn" :disabled="isSubmitting" @tap="handleAutoEval">
				<text class="iconfont icon-auto_awesome_mosaic"></text>
				<text class="btn-text">{{ isSubmitting ? '评教中...' : '一键评教' }}</text>
			</button>
		</view>

		<!-- 结果弹窗 -->
		<view v-if="resultVisible" class="result-overlay" @tap="resultVisible = false">
			<view class="result-modal" @tap.stop>
				<view class="result-header" :class="resultData.failed === 0 ? 'success' : 'warning'">
					<text class="iconfont" :class="resultData.failed === 0 ? 'icon-check_circle' : 'icon-info'"></text>
					<text class="result-title">{{ resultData.failed === 0 ? '评教完成' : '评教部分完成' }}</text>
				</view>
				<view class="result-body">
					<view class="result-stats">
						<view class="stat-item">
							<text class="stat-value success">{{ resultData.succeeded }}</text>
							<text class="stat-label">成功</text>
						</view>
						<view class="stat-divider"></view>
						<view class="stat-item">
							<text class="stat-value fail">{{ resultData.failed }}</text>
							<text class="stat-label">失败</text>
						</view>
					</view>
					<view class="result-message">{{ resultData.message }}</view>
				</view>
				<view class="result-footer">
					<button class="result-btn" @tap="resultVisible = false; handleRefresh()">确定</button>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { jwxtApi, type EvaluationItem } from '@/services/apiService';
import { recordAppUsage } from '@/utils/appUsage';

const isLoading = ref(false);
const isSubmitting = ref(false);
const evaluations = ref<EvaluationItem[]>([]);
const emptyMessage = ref('暂无待评教课程');

// 结果弹窗
const resultVisible = ref(false);
const resultData = ref({
	succeeded: 0,
	failed: 0,
	message: ''
});

// 颜色类
const colorClasses = ['blue', 'purple', 'red', 'orange', 'green'];
const getColorClass = (index: number) => colorClasses[index % colorClasses.length];

// 加载待评教列表
const loadEvaluations = async () => {
	isLoading.value = true;
	try {
		const res = await jwxtApi.getEvaluationPending();
		console.log('[Evaluation] Pending:', JSON.stringify(res));
		
		if (res.data?.success && res.data.evaluations) {
			evaluations.value = res.data.evaluations;
			if (evaluations.value.length === 0) {
				emptyMessage.value = '太棒了！没有待评教课程';
			}
		} else {
			evaluations.value = [];
			emptyMessage.value = res.error || '获取待评教列表失败';
		}
	} catch (error) {
		console.error('[Evaluation] Failed to load:', error);
		evaluations.value = [];
		emptyMessage.value = '获取失败，请稍后重试';
	} finally {
		isLoading.value = false;
	}
};

// 一键评教
const handleAutoEval = () => {
	uni.showModal({
		title: '一键评教',
		content: `确定要对 ${evaluations.value.length} 门课程进行自动评教吗？\n将为所有课程选择最高评价。`,
		success: async (res) => {
			if (res.confirm) {
				await doAutoEvaluate();
			}
		}
	});
};

// 执行自动评教
const doAutoEvaluate = async () => {
	isSubmitting.value = true;
	uni.showLoading({ title: '正在评教...', mask: true });
	
	try {
		const res = await jwxtApi.autoEvaluate();
		console.log('[Evaluation] Auto result:', JSON.stringify(res));
		
		uni.hideLoading();
		
		if (res.data?.success !== false) {
			resultData.value = {
				succeeded: res.data?.succeeded || 0,
				failed: res.data?.failed || 0,
				message: res.data?.message || '评教完成'
			};
			resultVisible.value = true;
		} else {
			uni.showToast({
				title: res.error || '评教失败',
				icon: 'none'
			});
		}
	} catch (error) {
		uni.hideLoading();
		console.error('[Evaluation] Auto failed:', error);
		uni.showToast({
			title: '评教失败，请稍后重试',
			icon: 'none'
		});
	} finally {
		isSubmitting.value = false;
	}
};

// 刷新
const handleRefresh = () => {
	loadEvaluations();
};

const goBack = () => {
	safeNavigateBack();
};

onMounted(() => {
	recordAppUsage('evaluation');
	loadEvaluations();
});
</script>

<style lang="scss" scoped>
.evaluation-page {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	background-color: $bg-light;
}

.header {
	position: sticky;
	top: 0;
	z-index: 50;
	background: rgba(248, 250, 252, 0.9);
	backdrop-filter: blur(20rpx);
	border-bottom: 2rpx solid rgba(0,0,0,0.05);
	padding: 24rpx 32rpx;
	padding-top: calc(var(--status-bar-height) + 16rpx);
}

.header-content {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.back-btn, .refresh-btn {
	width: 64rpx;
	height: 64rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	
	.iconfont {
		font-size: 36rpx;
		color: $primary;
	}
	
	&:active {
		background-color: rgba(0,0,0,0.05);
	}
}

.page-title {
	font-size: 36rpx;
	font-weight: 700;
	color: $text-primary;
}

.content {
	flex: 1;
	padding: 32rpx;
	padding-bottom: 200rpx;
}

.section-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-end;
	margin-bottom: 24rpx;
	padding: 0 8rpx;
}

.section-title {
	font-size: 36rpx;
	font-weight: 700;
	color: $text-primary;
}

.pending-badge {
	padding: 8rpx 16rpx;
	background-color: rgba(59, 130, 246, 0.1);
	border-radius: 12rpx;
}

.badge-text {
	font-size: 26rpx;
	font-weight: 500;
	color: $primary;
}

// 加载状态
.loading-container {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 128rpx 0;
}

.loading-text {
	font-size: 28rpx;
	color: $text-light;
}

// 空状态
.empty-container {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 128rpx 0;
}

.empty-icon {
	font-size: 96rpx;
	color: #4ade80;
	margin-bottom: 24rpx;
}

.empty-text {
	font-size: 28rpx;
	color: $text-light;
}

.courses-list {
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.course-card {
	background-color: #fff;
	border-radius: 32rpx;
	padding: 40rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
}

.card-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
}

.card-left {
	flex: 1;
	display: flex;
	flex-direction: column;
}

.pending-tag {
	display: inline-flex;
	padding: 8rpx 20rpx;
	background-color: rgba(249, 115, 22, 0.1);
	border-radius: 32rpx;
	border: 2rpx solid rgba(249, 115, 22, 0.2);
	margin-bottom: 16rpx;
	align-self: flex-start;
}

.tag-text {
	font-size: 24rpx;
	font-weight: 500;
	color: #f97316;
}

.course-title {
	font-size: 32rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 8rpx;
}

.course-dept {
	font-size: 24rpx;
	color: $text-secondary;
	font-weight: 500;
}

.course-icon {
	width: 80rpx;
	height: 80rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	
	.iconfont {
		font-size: 36rpx;
	}
	
	&.blue {
		background-color: rgba(59, 130, 246, 0.1);
		color: $primary;
	}
	
	&.purple {
		background-color: rgba(139, 92, 246, 0.1);
		color: #8b5cf6;
	}
	
	&.red {
		background-color: rgba(239, 68, 68, 0.1);
		color: $danger;
	}
	
	&.orange {
		background-color: rgba(249, 115, 22, 0.1);
		color: #f97316;
	}
	
	&.green {
		background-color: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
}

.card-divider {
	height: 2rpx;
	background-color: $border-color;
	margin: 24rpx 0;
}

.card-footer {
	display: flex;
	align-items: center;
	gap: 24rpx;
}

.teacher-icon {
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	background-color: rgba(59, 130, 246, 0.1);
	display: flex;
	align-items: center;
	justify-content: center;
	
	.iconfont {
		font-size: 32rpx;
		color: $primary;
	}
}

.prof-info {
	display: flex;
	flex-direction: column;
}

.prof-name {
	font-size: 28rpx;
	font-weight: 700;
	color: $text-primary;
}

.prof-role {
	font-size: 22rpx;
	color: $text-light;
}

.end-tip {
	display: flex;
	justify-content: center;
	padding: 48rpx 0;
}

.tip-text {
	font-size: 24rpx;
	color: $text-light;
	opacity: 0.6;
}

.bottom-action {
	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background: linear-gradient(to top, $bg-light 0%, $bg-light 80%, transparent 100%);
	padding: 40rpx 48rpx;
	padding-bottom: calc(32rpx + env(safe-area-inset-bottom));
}

.action-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 16rpx;
	width: 100%;
	height: 100rpx;
	background-color: $primary;
	border-radius: 32rpx;
	box-shadow: 0 16rpx 40rpx rgba(59, 130, 246, 0.3);
	border: none;
	
	.iconfont {
		font-size: 36rpx;
		color: #fff;
	}
	
	.btn-text {
		font-size: 32rpx;
		font-weight: 700;
		color: #fff;
	}
	
	&:active {
		transform: scale(0.98);
		opacity: 0.9;
	}
	
	&[disabled] {
		opacity: 0.6;
	}
}

// 结果弹窗
.result-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
	z-index: 100;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 48rpx;
}

.result-modal {
	width: 100%;
	max-width: 560rpx;
	background-color: #fff;
	border-radius: 32rpx;
	overflow: hidden;
}

.result-header {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 48rpx 40rpx 32rpx;
	
	.iconfont {
		font-size: 72rpx;
		margin-bottom: 16rpx;
	}
	
	&.success {
		background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
		
		.iconfont, .result-title {
			color: #fff;
		}
	}
	
	&.warning {
		background: linear-gradient(135deg, #f97316 0%, #ea580c 100%);
		
		.iconfont, .result-title {
			color: #fff;
		}
	}
}

.result-title {
	font-size: 32rpx;
	font-weight: 700;
}

.result-body {
	padding: 40rpx;
}

.result-stats {
	display: flex;
	justify-content: center;
	align-items: center;
	gap: 48rpx;
	margin-bottom: 24rpx;
}

.stat-item {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.stat-value {
	font-size: 48rpx;
	font-weight: 700;
	
	&.success {
		color: #22c55e;
	}
	
	&.fail {
		color: #ef4444;
	}
}

.stat-label {
	font-size: 24rpx;
	color: $text-light;
	margin-top: 8rpx;
}

.stat-divider {
	width: 2rpx;
	height: 48rpx;
	background-color: $border-color;
}

.result-message {
	text-align: center;
	font-size: 26rpx;
	color: $text-secondary;
}

.result-footer {
	padding: 0 40rpx 40rpx;
}

.result-btn {
	width: 100%;
	height: 88rpx;
	background-color: $primary;
	border-radius: 24rpx;
	border: none;
	font-size: 30rpx;
	font-weight: 600;
	color: #fff;
}
</style>
