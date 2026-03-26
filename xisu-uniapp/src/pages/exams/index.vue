<template>
	<view class="exams-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-content">
				<view class="back-btn" @tap="goBack">
					<text class="iconfont icon-arrow_back"></text>
				</view>
				<text class="page-title">考试安排</text>
				<view class="refresh-btn" @tap="handleRefresh">
					<text class="iconfont icon-arrow_forward"></text>
				</view>
			</view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- 学期选择 -->
			<view class="term-selector">
				<scroll-view class="term-scroll" scroll-x>
					<view class="term-list">
						<view 
							v-for="term in semesters" 
							:key="term.id"
							class="term-item"
							:class="{ active: currentSemesterId === term.id }"
							@tap="selectSemester(term.id)"
						>
							<text class="term-text">{{ formatSemesterName(term.name) }}</text>
						</view>
					</view>
				</scroll-view>
			</view>

			<!-- 加载状态 -->
			<view v-if="isLoading" class="loading-container">
				<text class="loading-text">加载中...</text>
			</view>

			<!-- 空状态 -->
			<view v-else-if="exams.length === 0" class="empty-container">
				<text class="iconfont icon-school empty-icon"></text>
				<text class="empty-text">{{ emptyMessage }}</text>
			</view>

			<!-- 考试卡片列表 -->
			<view v-else class="exams-list">
				<view 
					v-for="exam in exams" 
					:key="exam.id"
					class="exam-card"
					@tap="showExamDetail(exam)"
				>
					<view class="exam-badge" :class="getBadgeColor(exam)">
						<text class="badge-text">{{ getBadgeText(exam) }}</text>
					</view>
					
					<view class="exam-header">
						<text class="exam-title">{{ getCourseName(exam) }}</text>
						<view class="exam-code" v-if="getExamType(exam)">
							<text class="code-text">{{ getExamType(exam) }}</text>
						</view>
					</view>
					
					<view class="exam-details">
						<view class="detail-item">
							<text class="detail-label">考试时间</text>
							<view class="detail-content">
								<text class="iconfont icon-calendar_today"></text>
								<text class="detail-value">{{ getExamTime(exam) }}</text>
							</view>
						</view>
						<view class="detail-divider"></view>
						<view class="detail-item">
							<text class="detail-label">考试地点</text>
							<view class="detail-content">
								<text class="iconfont icon-map"></text>
								<text class="detail-value">{{ getLocation(exam) }}</text>
							</view>
							<text class="detail-sub" v-if="getSeat(exam)">座位号: {{ getSeat(exam) }}</text>
						</view>
					</view>
				</view>
			</view>

			<!-- 底部提示 -->
			<view v-if="exams.length > 0" class="end-tip">
				<text class="tip-text">共 {{ exams.length }} 场考试</text>
			</view>
		</scroll-view>

		<!-- 考试详情弹窗 -->
		<view v-if="detailVisible" class="detail-overlay" @tap="detailVisible = false">
			<view class="detail-modal" @tap.stop>
				<view class="detail-header">
					<text class="detail-title">{{ selectedExam ? getCourseName(selectedExam) : '' }}</text>
					<view class="detail-close" @tap="detailVisible = false">
						<text class="iconfont icon-close"></text>
					</view>
				</view>
				<view class="detail-body" v-if="selectedExam">
					<view class="detail-info-list">
						<view 
							v-for="(item, index) in getDetailItems(selectedExam)" 
							:key="index"
							class="detail-info-item"
						>
							<text class="detail-info-label">{{ item.label }}</text>
							<text class="detail-info-value">{{ item.value }}</text>
						</view>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { jwxtApi, type ExamItem, type SemesterInfo } from '@/services/apiService';
import { recordAppUsage } from '@/utils/appUsage';

const isLoading = ref(false);
const currentSemesterId = ref('');
const semesters = ref<SemesterInfo[]>([]);
const exams = ref<ExamItem[]>([]);
const emptyMessage = ref('请选择学期查看考试安排');

// 详情弹窗
const detailVisible = ref(false);
const selectedExam = ref<ExamItem | null>(null);

// 格式化学期名称
const formatSemesterName = (name: string) => {
	const match = name.match(/(\d{4})-(\d{4})学年第([一二])学期/);
	if (match) {
		return `${match[1].slice(2)}-${match[2].slice(2)}${match[3] === '一' ? '秋' : '春'}`;
	}
	return name;
};

// 提取课程名称
const getCourseName = (exam: ExamItem): string => {
	if (exam.course_name) return exam.course_name;
	for (const key of ['课程名称', '课程', '科目']) {
		if (exam[key]) return String(exam[key]);
	}
	return '未知课程';
};

// 提取考试时间
const getExamTime = (exam: ExamItem): string => {
	if (exam.exam_time) return exam.exam_time;
	for (const key of ['考试时间', '考试安排', '时间', '考试日期']) {
		if (exam[key] && exam[key] !== '时间未安排') return String(exam[key]);
	}
	return '时间待定';
};

// 提取考试地点
const getLocation = (exam: ExamItem): string => {
	if (exam.location) return exam.location;
	for (const key of ['考试地点', '地点', '教室', '考场']) {
		if (exam[key] && exam[key] !== '地点未安排') return String(exam[key]);
	}
	return '地点待定';
};

// 提取座位号
const getSeat = (exam: ExamItem): string => {
	if (exam.seat) return exam.seat;
	for (const key of ['座位号', '座位', '座号']) {
		if (exam[key]) return String(exam[key]);
	}
	return '';
};

// 提取考试类型
const getExamType = (exam: ExamItem): string => {
	if (exam.exam_type) return exam.exam_type;
	for (const key of ['考试类别', '考试类型', '类型', '考核方式']) {
		if (exam[key]) return String(exam[key]);
	}
	return '';
};

// 计算剩余天数
const getDaysLeft = (exam: ExamItem): number | null => {
	const examTime = exam.exam_time || exam.exam_date || exam['考试日期'] || exam['考试时间'];
	if (!examTime) return null;
	
	// 尝试提取日期
	const dateMatch = String(examTime).match(/(\d{4})[年\-\/]?(\d{1,2})[月\-\/]?(\d{1,2})/);
	if (dateMatch) {
		const examDate = new Date(parseInt(dateMatch[1]), parseInt(dateMatch[2]) - 1, parseInt(dateMatch[3]));
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const diffTime = examDate.getTime() - today.getTime();
		const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
		return diffDays;
	}
	return null;
};

// 获取徽章文本
const getBadgeText = (exam: ExamItem): string => {
	const daysLeft = getDaysLeft(exam);
	if (daysLeft === null) return '时间待定';
	if (daysLeft < 0) return '已结束';
	if (daysLeft === 0) return '今天';
	if (daysLeft === 1) return '明天';
	return `剩余 ${daysLeft} 天`;
};

// 获取徽章颜色
const getBadgeColor = (exam: ExamItem): string => {
	const daysLeft = getDaysLeft(exam);
	if (daysLeft === null) return 'gray';
	if (daysLeft < 0) return 'gray';
	if (daysLeft <= 3) return 'red';
	if (daysLeft <= 7) return 'orange';
	return 'blue';
};

// 显示考试详情
const showExamDetail = (exam: ExamItem) => {
	selectedExam.value = exam;
	detailVisible.value = true;
};

// 获取详情项目
const getDetailItems = (exam: ExamItem) => {
	const items: Array<{ label: string; value: string }> = [];
	
	// 定义要显示的字段
	const fieldOrder = [
		{ key: 'exam_time', label: '考试时间' },
		{ key: '考试日期', label: '考试日期' },
		{ key: '考试安排', label: '考试安排' },
		{ key: 'location', label: '考试地点' },
		{ key: '考试地点', label: '考试地点' },
		{ key: 'seat', label: '座位号' },
		{ key: '座位号', label: '座位号' },
		{ key: 'exam_type', label: '考试类型' },
		{ key: '考试类别', label: '考试类别' },
		{ key: 'status', label: '考试情况' },
		{ key: '考试情况', label: '考试情况' },
	];
	
	const addedLabels = new Set<string>();
	for (const field of fieldOrder) {
		const value = exam[field.key];
		if (value !== undefined && value !== '' && !addedLabels.has(field.label)) {
			items.push({ label: field.label, value: String(value) });
			addedLabels.add(field.label);
		}
	}
	
	return items;
};

// 加载学期列表
const loadSemesters = async () => {
	try {
		const res = await jwxtApi.getSemesters();
		if (res.data?.semesters) {
			semesters.value = res.data.semesters;
			const current = semesters.value.find(s => s.current);
			if (current) {
				currentSemesterId.value = current.id;
			} else if (res.data.current_semester) {
				currentSemesterId.value = res.data.current_semester;
			} else if (semesters.value.length > 0) {
				currentSemesterId.value = semesters.value[0].id;
			}
		}
	} catch (error) {
		console.error('[Exams] Failed to load semesters:', error);
	}
};

// 加载考试安排
const loadExams = async () => {
	if (!currentSemesterId.value) {
		emptyMessage.value = '请选择学期查看考试安排';
		return;
	}
	
	isLoading.value = true;
	try {
		const res = await jwxtApi.getExams(currentSemesterId.value);
		console.log('[Exams] Response:', JSON.stringify(res));
		
		if (res.data?.success && res.data.exams) {
			exams.value = res.data.exams;
			if (exams.value.length === 0) {
				emptyMessage.value = '本学期暂无考试安排';
			}
		} else {
			exams.value = [];
			emptyMessage.value = res.error || '获取考试安排失败';
		}
	} catch (error) {
		console.error('[Exams] Failed to load:', error);
		exams.value = [];
		emptyMessage.value = '获取考试安排失败，请稍后重试';
	} finally {
		isLoading.value = false;
	}
};

// 选择学期
const selectSemester = (semesterId: string) => {
	if (currentSemesterId.value !== semesterId) {
		currentSemesterId.value = semesterId;
		loadExams();
	}
};

// 刷新
const handleRefresh = () => {
	loadExams();
};

const goBack = () => {
	safeNavigateBack();
};

onMounted(async () => {
	recordAppUsage('exams');
	await loadSemesters();
	if (currentSemesterId.value) {
		loadExams();
	}
});
</script>

<style lang="scss" scoped>
.exams-page {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
	background-color: $bg-light;
}

.header {
	position: sticky;
	top: 0;
	z-index: 50;
	background: rgba(248, 250, 252, 0.95);
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
}

// 学期选择器
.term-selector {
	margin-bottom: 32rpx;
}

.term-scroll {
	white-space: nowrap;
}

.term-list {
	display: inline-flex;
	gap: 24rpx;
	padding-bottom: 8rpx;
}

.term-item {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	padding: 16rpx 32rpx;
	background-color: #fff;
	border-radius: 32rpx;
	border: 2rpx solid $border-color;
	
	&.active {
		background-color: $primary;
		border-color: $primary;
		
		.term-text {
			color: #fff;
		}
	}
}

.term-text {
	font-size: 26rpx;
	color: $text-secondary;
	white-space: nowrap;
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
	color: $text-light;
	opacity: 0.5;
	margin-bottom: 24rpx;
}

.empty-text {
	font-size: 28rpx;
	color: $text-light;
}

// 考试列表
.exams-list {
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.exam-card {
	position: relative;
	background-color: #fff;
	border-radius: 32rpx;
	padding: 40rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
	overflow: hidden;
}

.exam-badge {
	position: absolute;
	top: 0;
	right: 0;
	padding: 12rpx 32rpx;
	border-bottom-left-radius: 32rpx;
	border-top-right-radius: 32rpx;
	
	&.red {
		background-color: rgba(239, 68, 68, 0.1);
		
		.badge-text {
			color: $danger;
		}
	}
	
	&.orange {
		background-color: rgba(249, 115, 22, 0.1);
		
		.badge-text {
			color: #f97316;
		}
	}
	
	&.blue {
		background-color: rgba(59, 130, 246, 0.1);
		
		.badge-text {
			color: $primary;
		}
	}
	
	&.gray {
		background-color: rgba(0, 0, 0, 0.05);
		
		.badge-text {
			color: $text-light;
		}
	}
}

.badge-text {
	font-size: 24rpx;
	font-weight: 700;
}

.exam-header {
	margin-bottom: 32rpx;
	padding-right: 140rpx;
}

.exam-title {
	font-size: 32rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 12rpx;
	line-height: 1.4;
}

.exam-code {
	display: inline-block;
	padding: 8rpx 16rpx;
	background-color: rgba(0,0,0,0.03);
	border: 2rpx solid rgba(0,0,0,0.04);
	border-radius: 12rpx;
}

.code-text {
	font-size: 24rpx;
	color: $text-secondary;
}

.exam-details {
	display: flex;
	gap: 24rpx;
}

.detail-item {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: 8rpx;
}

.detail-label {
	font-size: 24rpx;
	color: $text-light;
	font-weight: 500;
}

.detail-content {
	display: flex;
	align-items: center;
	gap: 8rpx;
	
	.iconfont {
		font-size: 28rpx;
		color: $primary;
	}
}

.detail-value {
	font-size: 26rpx;
	font-weight: 600;
	color: $text-primary;
}

.detail-sub {
	font-size: 22rpx;
	color: $text-light;
	padding-left: 36rpx;
}

.detail-divider {
	width: 2rpx;
	background-color: $border-color;
}

.end-tip {
	text-align: center;
	padding: 48rpx 0;
}

.tip-text {
	font-size: 24rpx;
	color: $text-light;
	opacity: 0.6;
}

// 详情弹窗
.detail-overlay {
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

.detail-modal {
	width: 100%;
	max-width: 600rpx;
	background-color: #fff;
	border-radius: 32rpx;
	overflow: hidden;
}

.detail-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx 40rpx;
	background: linear-gradient(135deg, $primary 0%, #2563EB 100%);
}

.detail-title {
	font-size: 32rpx;
	font-weight: 700;
	color: #fff;
	flex: 1;
	padding-right: 16rpx;
}

.detail-close {
	width: 56rpx;
	height: 56rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	background-color: rgba(255, 255, 255, 0.2);
	
	.iconfont {
		font-size: 28rpx;
		color: #fff;
	}
}

.detail-body {
	padding: 32rpx 40rpx;
}

.detail-info-list {
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.detail-info-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding-bottom: 24rpx;
	border-bottom: 2rpx solid $border-color;
	
	&:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}
}

.detail-info-label {
	font-size: 28rpx;
	color: $text-light;
}

.detail-info-value {
	font-size: 28rpx;
	font-weight: 500;
	color: $text-primary;
	text-align: right;
}
</style>
