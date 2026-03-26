<template>
	<view class="grades-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-content">
				<view class="back-btn" @tap="goBack">
					<text class="iconfont icon-arrow_back"></text>
				</view>
				<text class="page-title">成绩查询</text>
				<view class="refresh-btn" @tap="handleRefresh">
					<text class="iconfont icon-arrow_forward"></text>
				</view>
			</view>
		</view>

		<scroll-view class="content" scroll-y>
			<!-- 统计卡片 -->
			<view class="gpa-card">
				<view class="gpa-bg-circle top"></view>
				<view class="gpa-bg-circle bottom"></view>
				<view class="gpa-content">
					<view class="gpa-item">
						<text class="gpa-label">及格门数</text>
						<text class="gpa-value pass">{{ passedCount }}</text>
					</view>
					<view class="gpa-divider"></view>
					<view class="gpa-item">
						<text class="gpa-label">挂科门数</text>
						<text class="gpa-value fail">{{ failedCount }}</text>
					</view>
				</view>
				<view class="gpa-footer">
					<text class="gpa-footer-text">平均分 {{ statistics.average_score ?? '--' }} | 已获 {{ passedCredits }} 学分</text>
				</view>
			</view>

			<!-- 学期选择 -->
			<view class="term-selector">
				<scroll-view class="term-scroll" scroll-x>
					<view class="term-list">
						<view 
							v-for="(term, index) in semesters" 
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
			<view v-else-if="grades.length === 0" class="empty-container">
				<text class="iconfont icon-school empty-icon"></text>
				<text class="empty-text">{{ emptyMessage }}</text>
			</view>

			<!-- 成绩列表 -->
			<view v-else class="grades-list">
				<view 
					v-for="(grade, index) in grades" 
					:key="grade.id"
					class="grade-card"
					@tap="showGradeDetail(grade)"
				>
					<view class="grade-indicator" :class="getColorClass(index)"></view>
					<view class="grade-main">
						<view class="grade-icon" :class="getColorClass(index)">
							<text class="iconfont icon-menu_book"></text>
						</view>
						<view class="grade-info">
							<text class="grade-title">{{ getCourseName(grade) }}</text>
							<view class="grade-meta">
								<view class="grade-type">{{ getCourseType(grade) }}</view>
								<text class="grade-credit">{{ getCredit(grade) }} 学分</text>
							</view>
						</view>
					</view>
					<view class="grade-score">
						<text class="score-value" :class="getColorClass(index) + '-text'">{{ getScore(grade) }}</text>
						<text class="score-gpa">{{ getGradePoint(grade) }}</text>
					</view>
				</view>
			</view>

			<!-- 底部占位 -->
			<view class="bottom-spacer"></view>
		</scroll-view>
		
		<!-- 成绩详情弹窗 -->
		<view v-if="detailVisible" class="detail-overlay" @tap="detailVisible = false">
			<view class="detail-modal" @tap.stop>
				<view class="detail-header">
					<text class="detail-title">成绩详情</text>
					<view class="detail-close" @tap="detailVisible = false">
						<text class="iconfont icon-close">✕</text>
					</view>
				</view>
				<view class="detail-content">
					<view class="detail-course-name">{{ getCourseName(selectedGrade!) }}</view>
					<view class="detail-score-display">
						<text class="detail-score-value">{{ getScore(selectedGrade!) }}</text>
						<text class="detail-score-label">总评成绩</text>
					</view>
					<view class="detail-info-list">
						<view v-for="(item, key) in getDetailItems(selectedGrade!)" :key="key" class="detail-info-item">
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
import { ref, onMounted, computed } from 'vue';
import { safeNavigateBack } from '@/utils/navigation';
import { jwxtApi, type GradeItem, type GradeStatistics, type SemesterInfo } from '@/services/apiService';
import { recordAppUsage } from '@/utils/appUsage';

const isLoading = ref(false);
const currentSemesterId = ref('');
const semesters = ref<SemesterInfo[]>([]);
const grades = ref<GradeItem[]>([]);
const statistics = ref<GradeStatistics>({
	total_courses: 0,
	average_score: null,
	weighted_average: null,
	total_credits: 0,
	grade_distribution: {},
});
const emptyMessage = ref('请选择学期查看成绩');

// 计算及格和挂科门数
const passedCount = computed(() => {
	return grades.value.filter(grade => {
		const score = getScore(grade);
		// 处理数字成绩
		const numScore = parseFloat(score);
		if (!isNaN(numScore)) return numScore >= 60;
		// 处理文字成绩
		return ['及格', '中', '良', '优', '合格', '通过'].includes(score);
	}).length;
});

const failedCount = computed(() => {
	return grades.value.filter(grade => {
		const score = getScore(grade);
		// 处理数字成绩
		const numScore = parseFloat(score);
		if (!isNaN(numScore)) return numScore < 60;
		// 处理文字成绩
		return ['不及格', '不合格', '未通过'].includes(score);
	}).length;
});

// 计算及格课程的总学分
const passedCredits = computed(() => {
	return grades.value.filter(grade => {
		const score = getScore(grade);
		const numScore = parseFloat(score);
		if (!isNaN(numScore)) return numScore >= 60;
		return ['及格', '中', '良', '优', '合格', '通过'].includes(score);
	}).reduce((total, grade) => {
		const credit = parseFloat(getCredit(grade));
		return total + (isNaN(credit) ? 0 : credit);
	}, 0);
});

// 颜色类
const colorClasses = ['pastel-blue', 'pastel-pink', 'pastel-purple', 'pastel-orange', 'pastel-green', 'pastel-teal'];
const getColorClass = (index: number) => colorClasses[index % colorClasses.length];

// 格式化学期名称
const formatSemesterName = (name: string) => {
	const match = name.match(/(\d{4})-(\d{4})学年第([一二])学期/);
	if (match) {
		return `${match[1].slice(2)}-${match[2].slice(2)}${match[3] === '一' ? '秋' : '春'}`;
	}
	return name;
};

// 从动态字段中提取课程名称
const getCourseName = (grade: GradeItem): string => {
	for (const key of ['课程名称', '课程', '名称', 'courseName']) {
		if (grade[key]) return String(grade[key]);
	}
	return '未知课程';
};

// 提取课程类型
const getCourseType = (grade: GradeItem): string => {
	for (const key of ['课程类别', '类别', '修读类别', 'type']) {
		if (grade[key]) return String(grade[key]);
	}
	return '';
};

// 提取学分
const getCredit = (grade: GradeItem): string => {
	for (const key of ['学分', 'credit']) {
		if (grade[key] !== undefined) return String(grade[key]);
	}
	return '0';
};

// 提取成绩
const getScore = (grade: GradeItem): string => {
	for (const key of ['最终', '总评成绩', '成绩', '总评', 'score']) {
		if (grade[key] !== undefined) return String(grade[key]);
	}
	return '--';
};

// 提取绩点
const getGradePoint = (grade: GradeItem): string => {
	for (const key of ['绩点', 'gradePoint', 'gpa']) {
		if (grade[key] !== undefined) return String(grade[key]);
	}
	// 根据成绩计算绩点
	const score = parseFloat(getScore(grade));
	if (!isNaN(score)) {
		if (score >= 90) return '4.0';
		if (score >= 85) return '3.7';
		if (score >= 80) return '3.3';
		if (score >= 75) return '3.0';
		if (score >= 70) return '2.7';
		if (score >= 65) return '2.3';
		if (score >= 60) return '2.0';
		return '0';
	}
	return '--';
};

// 加载学期列表
const loadSemesters = async () => {
	try {
		const res = await jwxtApi.getSemesters();
		if (res.data?.semesters) {
			semesters.value = res.data.semesters;
			// 设置当前学期
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
		console.error('[Grades] Failed to load semesters:', error);
	}
};

// 加载成绩
const loadGrades = async () => {
	if (!currentSemesterId.value) {
		emptyMessage.value = '请选择学期查看成绩';
		return;
	}
	
	isLoading.value = true;
	try {
		const res = await jwxtApi.getGrades(currentSemesterId.value);
		console.log('[Grades] Response:', JSON.stringify(res));
		
		if (res.data?.success) {
			grades.value = res.data.grades || [];
			if (res.data.statistics) {
				statistics.value = res.data.statistics;
			}
			if (grades.value.length === 0) {
				emptyMessage.value = '该学期暂无成绩记录';
			}
		} else {
			grades.value = [];
			emptyMessage.value = res.error || res.data?.message || '获取成绩失败';
		}
	} catch (error: unknown) {
		console.error('[Grades] Error:', error);
		grades.value = [];
		emptyMessage.value = error instanceof Error ? error.message : '网络错误，请稍后重试';
	} finally {
		isLoading.value = false;
	}
};

// 选择学期
const selectSemester = (semesterId: string) => {
	if (currentSemesterId.value === semesterId) return;
	currentSemesterId.value = semesterId;
	loadGrades();
};

// 刷新
const handleRefresh = () => {
	loadGrades();
};

// 成绩详情弹窗
const detailVisible = ref(false);
const selectedGrade = ref<GradeItem | null>(null);

const showGradeDetail = (grade: GradeItem) => {
	selectedGrade.value = grade;
	detailVisible.value = true;
};

// 获取详情项目列表
const getDetailItems = (grade: GradeItem) => {
	const items: Array<{ label: string; value: string }> = [];
	
	// 定义要显示的字段（按顺序）
	const fieldOrder = [
		{ key: '学年学期', label: '学年学期' },
		{ key: '课程代码', label: '课程代码' },
		{ key: '课程序号', label: '课程序号' },
		{ key: '课程类别', label: '课程类别' },
		{ key: '学分', label: '学分' },
		{ key: '总评成绩', label: '总评成绩' },
		{ key: '补考成绩', label: '补考成绩' },
		{ key: '最终', label: '最终成绩' },
		{ key: '绩点', label: '绩点' },
	];
	
	for (const field of fieldOrder) {
		const value = grade[field.key];
		if (value !== undefined && value !== '') {
			items.push({ label: field.label, value: String(value) });
		}
	}
	
	return items;
};

const goBack = () => {
	safeNavigateBack();
};

onMounted(async () => {
	recordAppUsage('grades');
	await loadSemesters();
	if (currentSemesterId.value) {
		loadGrades();
	}
});
</script>

<style lang="scss" scoped>
.grades-page {
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

.page-title {
	font-size: 40rpx;
	font-weight: 700;
	color: $text-primary;
}

.refresh-btn {
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

.content {
	flex: 1;
	padding: 32rpx;
}

.gpa-card {
	position: relative;
	background: linear-gradient(135deg, $primary 0%, #2563EB 100%);
	border-radius: 48rpx;
	padding: 48rpx;
	margin-bottom: 32rpx;
	overflow: hidden;
	box-shadow: 0 16rpx 48rpx rgba(59, 130, 246, 0.2);
}

.gpa-bg-circle {
	position: absolute;
	border-radius: 50%;
	background-color: rgba(255,255,255,0.1);
	filter: blur(40rpx);
	
	&.top {
		width: 256rpx;
		height: 256rpx;
		top: -64rpx;
		right: -64rpx;
	}
	
	&.bottom {
		width: 192rpx;
		height: 192rpx;
		bottom: -64rpx;
		left: -64rpx;
	}
}

.gpa-content {
	position: relative;
	z-index: 10;
	display: flex;
	justify-content: space-around;
	align-items: center;
}

.gpa-item {
	text-align: center;
}

.gpa-label {
	display: block;
	font-size: 26rpx;
	color: rgba(255,255,255,0.85);
	margin-bottom: 12rpx;
}

.gpa-value {
	font-size: 64rpx;
	font-weight: 700;
	color: #fff;
	
	&.pass {
		color: #4ade80;
	}
	
	&.fail {
		color: #f87171;
	}
}

.gpa-divider {
	width: 2rpx;
	height: 80rpx;
	background-color: rgba(255,255,255,0.2);
}

.gpa-footer {
	position: relative;
	z-index: 10;
	margin-top: 24rpx;
	text-align: center;
}

.gpa-footer-text {
	font-size: 24rpx;
	color: rgba(255,255,255,0.7);
}

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
	flex-shrink: 0;
	padding: 20rpx 40rpx;
	border-radius: 32rpx;
	background-color: #fff;
	border: 2rpx solid $border-color;
	box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.05);
	
	&.active {
		background-color: $text-primary;
		border-color: $text-primary;
		box-shadow: 0 8rpx 24rpx rgba(0,0,0,0.2);
		
		.term-text {
			color: #fff;
		}
	}
}

.term-text {
	font-size: 28rpx;
	font-weight: 600;
	color: $text-secondary;
	white-space: nowrap;
}

.grades-list {
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.grade-card {
	position: relative;
	display: flex;
	align-items: center;
	justify-content: space-between;
	background-color: #fff;
	border-radius: 32rpx;
	padding: 32rpx;
	padding-left: 40rpx;
	box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.04);
	border: 2rpx solid rgba(0,0,0,0.04);
	overflow: hidden;
}

.grade-indicator {
	position: absolute;
	left: 0;
	top: 0;
	bottom: 0;
	width: 12rpx;
	
	&.pastel-blue { background-color: $pastel-blue-text; }
	&.pastel-pink { background-color: $pastel-pink-text; }
	&.pastel-purple { background-color: $pastel-purple-text; }
	&.pastel-orange { background-color: $pastel-orange-text; }
	&.pastel-green { background-color: $pastel-green-text; }
}

.grade-main {
	display: flex;
	align-items: flex-start;
	gap: 24rpx;
}

.grade-icon {
	width: 96rpx;
	height: 96rpx;
	border-radius: 32rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	
	.icon {
		font-size: 40rpx;
	}
	
	&.pastel-blue { background-color: $pastel-blue; color: $pastel-blue-text; }
	&.pastel-pink { background-color: $pastel-pink; color: $pastel-pink-text; }
	&.pastel-purple { background-color: $pastel-purple; color: $pastel-purple-text; }
	&.pastel-orange { background-color: $pastel-orange; color: $pastel-orange-text; }
	&.pastel-green { background-color: $pastel-green; color: $pastel-green-text; }
}

.grade-info {
	flex: 1;
}

.grade-title {
	font-size: 32rpx;
	font-weight: 700;
	color: $text-primary;
	margin-bottom: 12rpx;
}

.grade-meta {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.grade-type {
	padding: 6rpx 16rpx;
	background-color: rgba(0,0,0,0.04);
	border-radius: 8rpx;
	font-size: 24rpx;
	color: $text-secondary;
}

.grade-credit {
	font-size: 24rpx;
	color: $text-secondary;
}

.grade-score {
	display: flex;
	flex-direction: column;
	align-items: flex-end;
}

.score-value {
	font-size: 48rpx;
	font-weight: 700;
	
	&.pastel-blue-text { color: $pastel-blue-text; }
	&.pastel-pink-text { color: $pastel-pink-text; }
	&.pastel-purple-text { color: $pastel-purple-text; }
	&.pastel-orange-text { color: $pastel-orange-text; }
	&.pastel-green-text { color: $pastel-green-text; }
}

.score-gpa {
	font-size: 24rpx;
	color: $text-light;
	margin-top: 4rpx;
}

.loading-container {
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 100rpx 0;
}

.loading-text {
	font-size: 28rpx;
	color: $text-secondary;
}

.empty-container {
	display: flex;
	flex-direction: column;
	justify-content: center;
	align-items: center;
	padding: 120rpx 0;
}

.empty-icon {
	font-size: 120rpx;
	color: $text-light;
	margin-bottom: 24rpx;
}

.empty-text {
	font-size: 28rpx;
	color: $text-secondary;
}

.pastel-teal {
	background-color: rgba(20, 184, 166, 0.15);
	
	.iconfont {
		color: #14b8a6;
	}
}

.pastel-teal-text {
	color: #14b8a6;
}

.bottom-spacer {
	height: 100rpx;
}

// 成绩详情弹窗
.detail-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.5);
	z-index: 1000;
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
	box-shadow: 0 32rpx 96rpx rgba(0, 0, 0, 0.2);
}

.detail-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 32rpx 40rpx;
	border-bottom: 2rpx solid $border-color;
}

.detail-title {
	font-size: 36rpx;
	font-weight: 700;
	color: $text-primary;
}

.detail-close {
	width: 56rpx;
	height: 56rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	background-color: rgba(0, 0, 0, 0.05);
	
	.iconfont {
		font-size: 28rpx;
		color: $text-secondary;
	}
	
	&:active {
		background-color: rgba(0, 0, 0, 0.1);
	}
}

.detail-content {
	padding: 40rpx;
}

.detail-course-name {
	font-size: 36rpx;
	font-weight: 600;
	color: $text-primary;
	text-align: center;
	margin-bottom: 32rpx;
}

.detail-score-display {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 32rpx;
	background: linear-gradient(135deg, $primary 0%, #2563EB 100%);
	border-radius: 24rpx;
	margin-bottom: 32rpx;
}

.detail-score-value {
	font-size: 72rpx;
	font-weight: 700;
	color: #fff;
}

.detail-score-label {
	font-size: 24rpx;
	color: rgba(255, 255, 255, 0.8);
	margin-top: 8rpx;
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
	padding: 20rpx 0;
	border-bottom: 2rpx solid rgba(0, 0, 0, 0.04);
	
	&:last-child {
		border-bottom: none;
	}
}

.detail-info-label {
	font-size: 28rpx;
	color: $text-secondary;
}

.detail-info-value {
	font-size: 28rpx;
	font-weight: 500;
	color: $text-primary;
}
</style>
