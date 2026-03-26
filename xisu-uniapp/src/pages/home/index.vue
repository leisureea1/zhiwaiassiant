<template>
	<view class="home-page">
		<!-- 头部 -->
		<view class="header">
			<view class="header-top" :style="{ paddingRight: headerRightPadding }">
				<view class="header-left">
					<!-- 学期+周数 多列选择器 -->
					<picker 
						mode="multiSelector" 
						:range="multiPickerRange" 
						:value="multiPickerValue"
						@change="onMultiPickerChange"
						@columnchange="onColumnChange"
					>
						<view class="week-selector">
							<view class="week-info">
								<text class="week-text">{{ isVacation ? '假期中' : `第${currentWeek}周` }}</text>
								<text class="icon-arrow">▼</text>
							</view>
							<text class="week-hint">{{ currentSemesterShortName }} · {{ weekHintText }}</text>
						</view>
					</picker>
				</view>
				<view class="refresh-btn" @tap="handleRefresh">
					<text class="iconfont icon-arrow_forward"></text>
					<text class="refresh-text">刷新</text>
				</view>
			</view>

			<!-- 公告流动展示（周数选择器与星期栏之间） -->
			<view class="announcement-strip" @tap="handleTickerTap">
				<text class="announcement-label">公告</text>
				<view class="announcement-marquee">
					<view class="announcement-track" :class="{ 'is-empty': !hasTickerAnnouncements }">
						<text class="announcement-text">{{ tickerAnnouncementText }}</text>
						<text v-if="hasTickerAnnouncements" class="announcement-text">{{ tickerAnnouncementText }}</text>
					</view>
				</view>
				<text class="announcement-more">详情</text>
			</view>
			
			<!-- 日期头 -->
			<view class="date-header">
				<view class="time-col"></view>
				<view class="day-col" v-for="(day, index) in weekDays" :key="index">
					<text class="day-name">{{ day.name }}</text>
					<text class="day-date">{{ day.date }}</text>
				</view>
			</view>
		</view>

		<!-- 课程表网格 -->
		<view class="schedule-grid">
			<view class="grid-container">
				<!-- 时间列 -->
				<view class="time-sidebar">
					<view 
						v-for="(slot, index) in timeSlots" 
						:key="index"
						class="time-slot"
						:class="{ 'break-slot': slot.isBreak }"
					>
						<template v-if="!slot.isBreak">
							<text class="slot-num">{{ slot.num }}</text>
							<text class="slot-time">{{ slot.time }}</text>
						</template>
						<template v-else>
							<text class="break-label">{{ slot.label }}</text>
						</template>
					</view>
				</view>
				
				<!-- 课程区域 -->
				<view class="courses-area">
					<!-- 午休行 -->
					<view class="break-row lunch-break" :style="{ top: (4 / 11.2 * 100) + '%' }">
						<view class="break-content">
							<text class="iconfont icon-calendar_today"></text>
							<text class="break-text">午休 12:00-14:00</text>
						</view>
					</view>
					
					<!-- 晚休行 -->
					<view class="break-row dinner-break" :style="{ top: (8.6 / 11.2 * 100) + '%' }">
						<view class="break-content">
							<text class="iconfont icon-calendar_today"></text>
							<text class="break-text">晚休 18:00-19:10</text>
						</view>
					</view>

					<!-- 课程卡片 -->
					<view 
						v-for="(course, index) in courses" 
						:key="index"
						class="course-card"
						:class="course.colorClass"
						:style="getCourseStyle(course)"
					>
						<text class="course-name">{{ course.name }}</text>
						<text class="course-teacher">{{ course.teacher }}</text>
						<text class="course-room">{{ course.room }}</text>
					</view>
				</view>
			</view>
		</view>
		
		<!-- 自定义 TabBar -->
		<TabBar :current="0" />

		<!-- 公告弹窗 -->
		<view v-if="showAnnouncementPopup && currentPopupAnnouncement" class="announcement-overlay" @tap="closeAnnouncementPopup">
			<view class="announcement-popup" @tap.stop>
				<!-- 关闭按钮 -->
				<view class="popup-close" @tap="closeAnnouncementPopup">
					<text class="close-icon">✕</text>
				</view>
				<!-- 类型标签 -->
				<view class="popup-type-badge" :class="'type-' + currentPopupAnnouncement.type">
					<text class="badge-text">{{ getAnnouncementTypeLabel(currentPopupAnnouncement.type) }}</text>
				</view>
				<!-- 标题 -->
				<text class="popup-title">{{ currentPopupAnnouncement.title }}</text>
				<!-- 时间 -->
				<text class="popup-time">{{ formatDate(currentPopupAnnouncement.publishedAt || currentPopupAnnouncement.createdAt) }}</text>
				<!-- 内容 -->
				<scroll-view scroll-y class="popup-content">
					<text v-if="popupContentText" class="popup-content-text">{{ popupContentText }}</text>
					<view v-if="popupImageUrls.length" class="popup-image-list">
						<image
							v-for="url in popupImageUrls"
							:key="url"
							:src="url"
							class="popup-content-image"
							mode="widthFix"
						/>
					</view>
				</scroll-view>
				<!-- 底部按钮 -->
				<view class="popup-footer">
					<view v-if="popupAnnouncements.length > 1" class="popup-counter">
						<text class="counter-text">{{ popupIndex + 1 }} / {{ popupAnnouncements.length }}</text>
					</view>
					<view class="popup-btn" @tap="handlePopupAction">
						<text class="btn-text">{{ popupIndex < popupAnnouncements.length - 1 ? '下一条' : '我知道了' }}</text>
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import TabBar from '@/components/TabBar/index.vue';
import { jwxtApi, announcementApi, getAccessToken } from '@/services/apiService';
import { recordAppUsage } from '@/utils/appUsage';
import type { AnnouncementItem } from '@/services/apiService';

// 学期周数配置
const TOTAL_WEEKS = 18; // 学期总周数

// ============ 课程表本地缓存 ============
const COURSES_CACHE_KEY = 'courses_cache';
const SEMESTERS_CACHE_KEY = 'semesters_cache';
const LAST_SEMESTER_ID_KEY = 'last_semester_id';
const CACHE_DURATION = 24 * 60 * 60 * 1000; // 缓存有效期：24小时

interface CacheData<T> {
	data: T;
	timestamp: number;
	semesterId?: string;
}

// 保存课程缓存
const saveCoursesCache = (semesterId: string, courses: DisplayCourse[]) => {
	try {
		const cacheData: CacheData<DisplayCourse[]> = {
			data: courses,
			timestamp: Date.now(),
			semesterId,
		};
		uni.setStorageSync(`${COURSES_CACHE_KEY}_${semesterId}`, JSON.stringify(cacheData));
		console.log('[Cache] Courses saved for semester:', semesterId, 'count:', courses.length);
	} catch (e) {
		console.error('[Cache] Failed to save courses:', e);
	}
};

// 获取课程缓存
const getCoursesCache = (semesterId: string): DisplayCourse[] | null => {
	try {
		const cached = uni.getStorageSync(`${COURSES_CACHE_KEY}_${semesterId}`);
		if (cached) {
			const cacheData: CacheData<DisplayCourse[]> = JSON.parse(cached);
			// 检查缓存是否过期
			if (Date.now() - cacheData.timestamp < CACHE_DURATION) {
				console.log('[Cache] Courses loaded from cache for semester:', semesterId);
				return cacheData.data;
			} else {
				console.log('[Cache] Courses cache expired for semester:', semesterId);
			}
		}
	} catch (e) {
		console.error('[Cache] Failed to read courses cache:', e);
	}
	return null;
};

// 保存学期缓存
const saveSemestersCache = (semesterList: SemesterInfo[], currentId: string) => {
	try {
		const cacheData: CacheData<{ list: SemesterInfo[]; currentId: string }> = {
			data: { list: semesterList, currentId },
			timestamp: Date.now(),
		};
		uni.setStorageSync(SEMESTERS_CACHE_KEY, JSON.stringify(cacheData));
		if (currentId) {
			uni.setStorageSync(LAST_SEMESTER_ID_KEY, currentId);
		}
		console.log('[Cache] Semesters saved, count:', semesterList.length);
	} catch (e) {
		console.error('[Cache] Failed to save semesters:', e);
	}
};

// 获取学期缓存
const getSemestersCache = (): { list: SemesterInfo[]; currentId: string } | null => {
	try {
		const cached = uni.getStorageSync(SEMESTERS_CACHE_KEY);
		if (cached) {
			const cacheData: CacheData<{ list: SemesterInfo[]; currentId: string }> = JSON.parse(cached);
			// 学期数据缓存有效期更长（7天）
			if (Date.now() - cacheData.timestamp < 7 * 24 * 60 * 60 * 1000) {
				console.log('[Cache] Semesters loaded from cache');
				return cacheData.data;
			}
		}
	} catch (e) {
		console.error('[Cache] Failed to read semesters cache:', e);
	}
	return null;
};

// 清除课程缓存（用于刷新）
const clearCoursesCache = (semesterId?: string) => {
	try {
		if (semesterId) {
			uni.removeStorageSync(`${COURSES_CACHE_KEY}_${semesterId}`);
		} else {
			// 清除所有课程缓存
			const keys = uni.getStorageInfoSync().keys;
			keys.forEach((key: string) => {
				if (key.startsWith(COURSES_CACHE_KEY)) {
					uni.removeStorageSync(key);
				}
			});
		}
		console.log('[Cache] Courses cache cleared');
	} catch (e) {
		console.error('[Cache] Failed to clear cache:', e);
	}
};

// 根据学期名称解析学期开始日期
const parseSemesterStartDate = (semesterName: string): Date | null => {
	// 学期名称格式: "2025-2026学年第一学期" 或 "2025-2026学年第二学期"
	const match = semesterName.match(/(\d{4})-(\d{4})学年第([一二])学期/);
	if (match) {
		const startYear = parseInt(match[1]);
		const semester = match[3];
		
		if (semester === '一') {
			// 秋季学期（第一学期）：从第一年的9月1日开始
			return new Date(startYear, 8, 1);
		} else {
			// 春季学期（第二学期）：从第二年的3月2日开始
			return new Date(startYear + 1, 2, 2);
		}
	}
	return null;
};

// 获取当前选中学期的开始日期
const getSelectedSemesterStart = (): Date => {
	// 先尝试根据选中的学期名称解析
	const semester = semesters.value.find(s => s.id === currentSemesterId.value);
	if (semester) {
		const parsedDate = parseSemesterStartDate(semester.name);
		if (parsedDate) {
			return parsedDate;
		}
	}
	
	// 兜底：根据当前日期推断
	return getCurrentSemesterStart();
};

// 获取当前学期的开始日期（根据系统日期推断，用于计算实际周数）
const getCurrentSemesterStart = (): Date => {
	const now = new Date();
	const year = now.getFullYear();
	const month = now.getMonth() + 1;

	// 春季学期：3月2日开始
	if (month >= 3 && month <= 7) {
		return new Date(year, 2, 2); // 3月2日
	}
	// 秋季学期：9月1日开始
	else if (month >= 9 || month <= 1) {
		const semesterYear = month >= 9 ? year : year - 1;
		return new Date(semesterYear, 8, 1); // 9月1日
	}
	// 2月份期间（寒假）仍按上一年秋季学期计算
	else {
		return new Date(year - 1, 8, 1); // 上一年9月1日
	}
};

// 计算周数
const getWeekNumber = (semesterStart: Date, currentDate: Date = new Date()): number => {
	const diffTime = currentDate.getTime() - semesterStart.getTime();
	const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
	return Math.ceil(diffDays / 7);
};

// 检查是否在假期中
const isInVacation = (currentDate: Date = new Date()): boolean => {
	const semesterStart = getCurrentSemesterStart();
	const weekNumber = getWeekNumber(semesterStart, currentDate);
	return weekNumber < 1 || weekNumber > TOTAL_WEEKS;
};

// 获取假期祝福语
const getVacationGreeting = (currentDate: Date = new Date()): string => {
	const month = currentDate.getMonth() + 1;
	
	if (month >= 1 && month <= 2) {
		return '寒假快乐 🎉';
	} else if (month >= 7 && month <= 8) {
		return '暑假快乐 ☀️';
	} else {
		return '假期快乐 🎊';
	}
};

const currentWeek = ref(1);
const actualCurrentWeek = ref(1); // 实际当前周
const isVacation = ref(false);
const vacationGreeting = ref('');
const headerRightPadding = ref('0px');
const isLoading = ref(false);

// 学期相关状态
interface SemesterInfo {
	id: string;
	name: string;
	current?: boolean;
}
const semesters = ref<SemesterInfo[]>([]);
const currentSemesterId = ref('');
const currentSemesterIndex = computed(() => {
	const idx = semesters.value.findIndex(s => s.id === currentSemesterId.value);
	return idx >= 0 ? idx : 0;
});
const currentSemesterName = computed(() => {
	const semester = semesters.value.find(s => s.id === currentSemesterId.value);
	return semester?.name || '加载中...';
});
// 学期简称（用于显示）
const currentSemesterShortName = computed(() => {
	const name = currentSemesterName.value;
	// 从 "2025-2026学年第二学期" 提取 "25-26春" 或类似格式
	const match = name.match(/(\d{4})-(\d{4})学年第([一二])学期/);
	if (match) {
		const startYear = match[1].slice(2);
		const endYear = match[2].slice(2);
		const term = match[3] === '一' ? '秋' : '春';
		return `${startYear}-${endYear}${term}`;
	}
	return name.length > 8 ? name.slice(0, 8) + '...' : name;
});

// 多列选择器数据
const multiPickerRange = computed(() => {
	// 第一列：学期列表
	const semesterNames = semesters.value.map(s => {
		// 简化学期名称显示
		const match = s.name.match(/(\d{4})-(\d{4})学年第([一二])学期/);
		if (match) {
			return `${match[1].slice(2)}-${match[2].slice(2)}${match[3] === '一' ? '秋' : '春'}`;
		}
		return s.name;
	});
	
	// 第二列：周数列表
	const weekNames: string[] = [];
	for (let i = 1; i <= TOTAL_WEEKS; i++) {
		let label = `第${i}周`;
		if (!isVacation.value && i === actualCurrentWeek.value) {
			label += '(当前)';
		}
		weekNames.push(label);
	}
	
	return [semesterNames.length > 0 ? semesterNames : ['加载中...'], weekNames];
});

const multiPickerValue = computed(() => [currentSemesterIndex.value, currentWeek.value - 1]);

// 多列选择器确认
const onMultiPickerChange = async (e: any) => {
	const values = e.detail.value;
	const semesterIdx = values[0];
	const weekIdx = values[1];
	
	console.log('[Home] MultiPicker change:', semesterIdx, weekIdx);
	
	// 检查是否有变化
	const semester = semesters.value[semesterIdx];
	const semesterChanged = semester && semester.id !== currentSemesterId.value;
	const weekChanged = (weekIdx + 1) !== currentWeek.value;
	
	if (!semesterChanged && !weekChanged) {
		return; // 没有变化，不需要加载
	}
	
	// 显示加载提示
	uni.showLoading({ title: '加载中...', mask: true });
	
	// 更新学期
	if (semesterChanged) {
		currentSemesterId.value = semester.id;
		uni.setStorageSync(LAST_SEMESTER_ID_KEY, semester.id);
	}
	
	// 更新周数
	currentWeek.value = weekIdx + 1;
	updateWeekDays();
	
	try {
		// 切换学期时强制刷新
		await loadCourses(semesterChanged);
		uni.hideLoading();
	} catch (error) {
		uni.hideLoading();
		uni.showToast({ title: '加载失败', icon: 'none' });
	}
};

// 列变化时的回调（可用于联动）
const onColumnChange = (e: any) => {
	console.log('[Home] Column change:', e.detail.column, e.detail.value);
};

// 加载学期列表（优先使用缓存）
const loadSemesters = async () => {
	// 先尝试从缓存加载
	const cached = getSemestersCache();
	if (cached && cached.list.length > 0) {
		semesters.value = cached.list;
		currentSemesterId.value = cached.currentId;
		console.log('[Home] Loaded semesters from cache:', semesters.value.length);
		// 后台静默更新
		fetchSemestersFromServer();
		return;
	}
	
	await fetchSemestersFromServer();
};

// 从服务器获取学期列表
const fetchSemestersFromServer = async () => {
	try {
		const res = await jwxtApi.getSemesters();
		console.log('[Home] Semesters response:', res);

		// 兼容多种历史/迁移期返回格式
		const root = (res as any) || {};
		const level1 = root.data || {};
		const level2 = level1.data || {};
		const semesterPayload = (level1.semesters ? level1 : (level2.semesters ? level2 : null));

		if (root.success && semesterPayload?.semesters) {
			// 后端返回的已是结构化数据 { id, name, current? }
			semesters.value = semesterPayload.semesters;

			// 设置当前学期
			const current = semesters.value.find(s => s.current);
			if (current) {
				currentSemesterId.value = current.id;
			} else if (semesterPayload.current_semester) {
				currentSemesterId.value = semesterPayload.current_semester;
			} else if (semesterPayload.current_semester_id) {
				currentSemesterId.value = semesterPayload.current_semester_id;
			} else if (semesters.value.length > 0) {
				currentSemesterId.value = semesters.value[0].id;
			}

			// 保存到缓存
			saveSemestersCache(semesters.value, currentSemesterId.value);
			console.log('[Home] Loaded semesters:', semesters.value.length, 'current:', currentSemesterId.value);
		}
	} catch (error) {
		console.error('[Home] Failed to load semesters:', error);
	}
};

// 计算当前是第几周
const calculateCurrentWeek = (): number => {
	const semesterStart = getCurrentSemesterStart();
	const week = getWeekNumber(semesterStart);
	return Math.max(1, Math.min(week, TOTAL_WEEKS));
};

// 周数提示文字
const weekHintText = computed(() => {
	if (isVacation.value) {
		return vacationGreeting.value;
	}
	if (currentWeek.value === actualCurrentWeek.value) {
		return '点击切换';
	} else {
		return `实际第${actualCurrentWeek.value}周`;
	}
});

// 颜色类数组，相邻颜色视觉差距大，避免相邻课程难以区分
const colorClasses = ['color-sky', 'color-pink', 'color-emerald', 'color-indigo', 'color-amber', 'color-teal', 'color-yellow', 'color-purple'];
const courseColorMap = new Map<string, string>();

const getWeekDays = () => {
	// 显示当前实际日期所在周的日期
	const now = new Date();
	const weekStart = new Date(now);
	
	// 调整到本周周一
	const dayOfWeek = weekStart.getDay() || 7;
	weekStart.setDate(weekStart.getDate() - dayOfWeek + 1);
	
	const days = [];
	const dayNames = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
	for (let i = 0; i < 7; i++) {
		const date = new Date(weekStart);
		date.setDate(weekStart.getDate() + i);
		days.push({
			name: dayNames[i],
			date: `${date.getMonth() + 1}/${date.getDate()}`
		});
	}
	return days;
};

// 更新日期头
const updateWeekDays = () => {
	weekDays.value = getWeekDays();
};

const weekDays = ref<{name: string; date: string}[]>([]);

const timeSlots = ref([
	{ num: 1, time: '8:00\n8:50' },
	{ num: 2, time: '9:00\n9:50' },
	{ num: 3, time: '10:10\n11:00' },
	{ num: 4, time: '11:10\n12:00' },
	{ isBreak: true, label: '午' },
	{ num: 6, time: '14:00\n14:50' },
	{ num: 7, time: '15:00\n15:50' },
	{ num: 8, time: '16:10\n17:00' },
	{ num: 9, time: '17:10\n18:00' },
	{ isBreak: true, label: '晚' },
	{ num: 11, time: '19:10\n20:00' },
	{ num: 12, time: '20:10\n21:00' }
]);

interface DisplayCourse {
	name: string;
	teacher: string;
	room: string;
	day: number;
	start: number;
	span: number;
	colorClass: string;
	weeks: number[]; // 该课程在哪些周有课
}

// 所有课程数据
const allCourses = ref<DisplayCourse[]>([]);

// 根据当前周过滤显示的课程
const courses = computed(() => {
	return allCourses.value.filter(course => {
		const name = (course.name || '').trim().toLowerCase();
		const normalized = name.replace(/[\s：:]/g, '');
		if (
			normalized === '课程名称' ||
			normalized === '课程名' ||
			normalized === '课程' ||
			normalized === 'coursename' ||
			normalized.includes('课程名称') ||
			normalized.includes('coursename')
		) {
			return false;
		}

		// 如果没有周次信息，默认显示
		if (!course.weeks || course.weeks.length === 0) {
			return true;
		}
		return course.weeks.includes(currentWeek.value);
	});
});

// 获取课程颜色
const getCourseColor = (courseName: string): string => {
	if (!courseColorMap.has(courseName)) {
		const colorIndex = courseColorMap.size % colorClasses.length;
		courseColorMap.set(courseName, colorClasses[colorIndex]);
	}
	return courseColorMap.get(courseName)!;
};

// 解析周次字符串，如 "1-8周, 10-16周" 转为数组 [1,2,3,...,8,10,11,...,16]
const parseWeeks = (weeksStr: string): number[] => {
	if (!weeksStr) return [];
	
	const weeks: number[] = [];
	// 匹配 "1-8周" 或 "1周" 这样的格式
	const matches = weeksStr.match(/(\d+)(?:-(\d+))?周/g);
	
	if (matches) {
		for (const match of matches) {
			const rangeMatch = match.match(/(\d+)(?:-(\d+))?周/);
			if (rangeMatch) {
				const start = parseInt(rangeMatch[1]);
				const end = rangeMatch[2] ? parseInt(rangeMatch[2]) : start;
				for (let i = start; i <= end; i++) {
					if (!weeks.includes(i)) {
						weeks.push(i);
					}
				}
			}
		}
	}
	
	return weeks.sort((a, b) => a - b);
};

// 将API返回的课程数据转换为显示格式
const transformCourses = (apiCourses: Array<{
	name: string;
	teacher: string;
	classroom: string;
	weekday: number;
	startSection: number;
	endSection: number;
	weeks?: string;
}>): DisplayCourse[] => {
	return apiCourses.map(course => ({
		name: course.name,
		teacher: course.teacher,
		room: course.classroom,
		day: course.weekday,
		start: course.startSection,
		span: course.endSection - course.startSection + 1,
		colorClass: getCourseColor(course.name),
		weeks: parseWeeks(course.weeks || ''),
	}));
};

// 从后端获取课程数据
const fetchCoursesFromServer = async (semesterId: string, showError = true): Promise<boolean> => {
	try {
		const res = await jwxtApi.getCourses(semesterId === 'default' ? undefined : semesterId);
		console.log('[Home] Courses response:', JSON.stringify(res));
		
		if (res.success && res.data?.courses) {
			allCourses.value = transformCourses(res.data.courses);
			// 保存到缓存
			saveCoursesCache(semesterId, allCourses.value);
			console.log('[Home] Loaded courses from server:', allCourses.value.length);
			return true;
		} else if (res.error && showError) {
			console.error('[Home] Failed to load courses:', res.error);
			uni.showToast({ title: res.error, icon: 'none' });
		}
		return false;
	} catch (error) {
		console.error('[Home] Error loading courses:', error);
		const errorMsg = error instanceof Error ? error.message : '加载课程表失败';
		
		if (showError) {
			// 检测是否是教务系统未绑定的错误
			if (errorMsg.includes('绑定') || errorMsg.includes('教务')) {
				uni.showModal({
					title: '提示',
					content: '您需要先绑定教务系统账号才能查看课程表，是否现在去绑定？',
					confirmText: '去绑定',
					cancelText: '稍后',
					success: (result) => {
						if (result.confirm) {
							uni.navigateTo({ url: '/pages/profile/bind-jwxt' });
						}
					}
				});
			} else {
				uni.showToast({ title: errorMsg, icon: 'none' });
			}
		}
		return false;
	}
};

// 从后端加载课程表：先显示缓存，同时异步更新
const loadCourses = async (forceRefresh = false) => {
	// 检查是否已登录
	const token = getAccessToken();
	if (!token) {
		console.log('[Home] Not logged in, skip loading courses');
		return;
	}
	
	const lastSemesterId = uni.getStorageSync(LAST_SEMESTER_ID_KEY) || '';
	const semesterId = currentSemesterId.value || lastSemesterId || 'default';
	if (!currentSemesterId.value && semesterId !== 'default') {
		currentSemesterId.value = semesterId;
	}
	
	// 1. 先尝试从缓存加载并立即显示
	const cachedCourses = getCoursesCache(semesterId);
	const hasCache = cachedCourses && cachedCourses.length > 0;
	
	if (hasCache && !forceRefresh) {
		allCourses.value = cachedCourses;
		// 重新分配颜色
		allCourses.value.forEach(course => {
			course.colorClass = getCourseColor(course.name);
		});
		console.log('[Home] Displayed from cache:', allCourses.value.length, 'courses');
	}
	
	// 2. 同时异步从后端获取最新数据
	if (!hasCache) {
		// 没有缓存时显示加载状态
		isLoading.value = true;
	}
	
	try {
		// 异步获取最新数据（有缓存时静默更新，无缓存时显示错误）
		await fetchCoursesFromServer(semesterId, !hasCache);
	} finally {
		isLoading.value = false;
	}
};

// 计算课程位置样式 - 使用百分比适配flex布局
const totalUnits = 11.2;
const getCourseStyle = (course: DisplayCourse) => {
	const dayWidth = 14.28; // 百分比 (100% / 7 天)
	const left = (course.day - 1) * dayWidth;
	
	// 计算top位置百分比
	let topUnits = 0;
	for (let i = 1; i < course.start; i++) {
		if (i === 5 || i === 10) {
			topUnits += 0.6; // 休息行
		} else {
			topUnits += 1; // 普通行
		}
	}
	const top = (topUnits / totalUnits) * 100;
	
	// 计算高度百分比
	const height = (course.span / totalUnits) * 100;
	
	return {
		left: `${left}%`,
		top: `${top}%`,
		width: `${dayWidth - 0.8}%`,
		height: `calc(${height}% - 4rpx)`
	};
};

const handleMenu = () => {
	uni.showToast({ title: '菜单', icon: 'none' });
};

const handleRefresh = async () => {
	uni.showLoading({ title: '刷新中...', mask: true });
	try {
		// 清除当前学期的课程缓存
		const semesterId = currentSemesterId.value || 'default';
		clearCoursesCache(semesterId);
		
		// 使用刷新接口（会清除后端缓存重新获取）
		const res = await jwxtApi.refreshCourses(semesterId === 'default' ? undefined : semesterId);
		console.log('[Home] Refresh courses response:', JSON.stringify(res));
		
		if (res.success && res.data?.courses) {
			allCourses.value = transformCourses(res.data.courses);
			// 保存到缓存
			saveCoursesCache(semesterId, allCourses.value);
		}
		uni.hideLoading();
		uni.showToast({ title: '刷新成功', icon: 'success' });
	} catch (error) {
		uni.hideLoading();
		uni.showToast({ title: '刷新失败', icon: 'none' });
	}
};

// ============ 公告弹窗 ============
const POPUP_VIEWED_KEY = 'announcement_popup_viewed';
const showAnnouncementPopup = ref(false);
const popupAnnouncements = ref<AnnouncementItem[]>([]);
const popupIndex = ref(0);
const currentPopupAnnouncement = computed(() => popupAnnouncements.value[popupIndex.value] || null);
const popupRawContent = computed(() => currentPopupAnnouncement.value?.summary || currentPopupAnnouncement.value?.content || '');

const markdownImageRegex = /!\[[^\]]*\]\((https?:\/\/[^)\s]+)\)/g;
const htmlImageRegex = /<img[^>]*src=["'](https?:\/\/[^"']+)["'][^>]*>/g;
const plainImageUrlRegex = /(https?:\/\/\S+\.(?:png|jpe?g|gif|webp|bmp|svg))/gi;

const popupImageUrls = computed(() => {
	const content = popupRawContent.value;
	if (!content) return [];

	const urls: string[] = [];
	let match: RegExpExecArray | null;

	while ((match = markdownImageRegex.exec(content)) !== null) {
		if (match[1]) urls.push(match[1]);
	}
	while ((match = htmlImageRegex.exec(content)) !== null) {
		if (match[1]) urls.push(match[1]);
	}
	while ((match = plainImageUrlRegex.exec(content)) !== null) {
		if (match[1]) urls.push(match[1]);
	}

	return Array.from(new Set(urls));
});

const popupContentText = computed(() => {
	const content = popupRawContent.value;
	if (!content) return '';

	return content
		.replace(markdownImageRegex, '')
		.replace(htmlImageRegex, '')
		.replace(plainImageUrlRegex, '')
		.replace(/\n{3,}/g, '\n\n')
		.trim();
});

const tickerAnnouncements = ref<AnnouncementItem[]>([]);
const hasTickerAnnouncements = computed(() => tickerAnnouncements.value.length > 0);
const tickerAnnouncementText = computed(() => {
	if (!tickerAnnouncements.value.length) {
		return '暂无公告';
	}

	return tickerAnnouncements.value
		.map((item) => `${getAnnouncementTypeLabel(item.type)}：${item.title}，点击查看详情`)
		.join('   ·   ');
});

// 获取已查看的弹窗公告 ID 集合
const getViewedPopupIds = (): Set<string> => {
	try {
		const stored = uni.getStorageSync(POPUP_VIEWED_KEY);
		if (stored) {
			const data = JSON.parse(stored);
			return new Set(data);
		}
	} catch {}
	return new Set();
};

// 保存已查看的弹窗公告 ID
const saveViewedPopupId = (id: string) => {
	try {
		const viewedIds = getViewedPopupIds();
		viewedIds.add(id);
		// 只保留最近 100 条记录
		const arr = Array.from(viewedIds).slice(-100);
		uni.setStorageSync(POPUP_VIEWED_KEY, JSON.stringify(arr));
	} catch {}
};

// 获取公告类型标签
const getAnnouncementTypeLabel = (type: string): string => {
	const map: Record<string, string> = {
		NORMAL: '公告',
		IMPORTANT: '重要',
		URGENT: '紧急',
		MAINTENANCE: '维护',
	};
	return map[type] || '公告';
};

// 格式化日期
const formatDate = (dateStr: string): string => {
	if (!dateStr) return '';
	const date = new Date(dateStr);
	const y = date.getFullYear();
	const m = String(date.getMonth() + 1).padStart(2, '0');
	const d = String(date.getDate()).padStart(2, '0');
	return `${y}-${m}-${d}`;
};

// 加载弹窗公告
const loadPopupAnnouncements = async () => {
	const token = getAccessToken();
	if (!token) return;

	try {
		const res = await announcementApi.getList({ page: 1, pageSize: 10 });
		if (res && res.items) {
			const publishedItems = res.items.filter((item: AnnouncementItem) => item.status === 'PUBLISHED');

			// 顶部跑马灯展示最近公告
			tickerAnnouncements.value = publishedItems.slice(0, 8);

			// 筛选弹窗公告（isPopup = true）
			const popupItems = publishedItems.filter((item: AnnouncementItem) => item.isPopup);
			if (popupItems.length === 0) return;

			// 排除已在本地查看过的
			const viewedIds = getViewedPopupIds();
			const unviewed = popupItems.filter((item: AnnouncementItem) => !viewedIds.has(item.id));
			if (unviewed.length === 0) return;

			popupAnnouncements.value = unviewed;
			popupIndex.value = 0;
			showAnnouncementPopup.value = true;
		}
	} catch (err) {
		console.log('[Home] Failed to load popup announcements:', err);
		tickerAnnouncements.value = [];
	}
};

// 点击公告条查看详情
const handleTickerTap = () => {
	if (!tickerAnnouncements.value.length) {
		loadPopupAnnouncements();
		uni.showToast({ title: '正在刷新公告', icon: 'none' });
		return;
	}

	popupAnnouncements.value = [...tickerAnnouncements.value];
	popupIndex.value = 0;
	showAnnouncementPopup.value = true;
};

// 关闭弹窗
const closeAnnouncementPopup = () => {
	// 标记当前公告已读
	const current = currentPopupAnnouncement.value;
	if (current) {
		saveViewedPopupId(current.id);
	}
	showAnnouncementPopup.value = false;
};

// 弹窗按钮操作
const handlePopupAction = () => {
	const current = currentPopupAnnouncement.value;
	if (current) {
		saveViewedPopupId(current.id);
	}

	if (popupIndex.value < popupAnnouncements.value.length - 1) {
		// 下一条
		popupIndex.value++;
	} else {
		// 最后一条，关闭
		showAnnouncementPopup.value = false;
	}
};

onMounted(async () => {
	recordAppUsage('schedule');
	try {
		const menuButton = uni.getMenuButtonBoundingClientRect?.();
		const systemInfo = uni.getSystemInfoSync?.();
		if (menuButton && systemInfo?.windowWidth) {
			const rightSpace = systemInfo.windowWidth - menuButton.left;
			headerRightPadding.value = `${rightSpace + 12}px`;
		}
	} catch (error) {
		headerRightPadding.value = '0px';
	}
	
	// 检查是否在假期
	isVacation.value = isInVacation();
	if (isVacation.value) {
		vacationGreeting.value = getVacationGreeting();
		// 假期期间默认显示第1周
		currentWeek.value = 1;
		actualCurrentWeek.value = 1;
	} else {
		// 初始化当前周
		actualCurrentWeek.value = calculateCurrentWeek();
		currentWeek.value = actualCurrentWeek.value;
	}
	updateWeekDays();

	// 冷启动优先使用上次学期 + 本地课程缓存，避免空白等待
	try {
		const lastSemesterId = uni.getStorageSync(LAST_SEMESTER_ID_KEY) || '';
		if (lastSemesterId && !currentSemesterId.value) {
			currentSemesterId.value = lastSemesterId;
		}
	} catch (e) {
		console.log('[Home] Failed to read last semester id:', e);
	}

	const startupSemesterId = currentSemesterId.value;
	loadCourses();
	
	// 再同步学期并按最新学期校正课表
	await loadSemesters();
	if (currentSemesterId.value !== startupSemesterId || allCourses.value.length === 0) {
		loadCourses();
	}

	// 加载弹窗公告
	loadPopupAnnouncements();
});
</script>

<style lang="scss" scoped>
.home-page {
	display: flex;
	flex-direction: column;
	height: 100vh;
	background-color: $bg-light;
}

.header {
	flex-shrink: 0;
	background-color: #fff;
	padding: 24rpx;
	padding-top: calc(var(--status-bar-height) + 48rpx);
	box-shadow: 0 2rpx 10rpx rgba(0,0,0,0.05);
	z-index: 30;
}

.header-top {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 24rpx;
}

.announcement-strip {
	display: flex;
	align-items: center;
	gap: 12rpx;
	height: 68rpx;
	padding: 0 16rpx;
	margin-bottom: 20rpx;
	border-radius: 16rpx;
	background: linear-gradient(90deg, #fff7cc 0%, #ffe38a 100%);
	border: 1rpx solid #f5cc5d;
	box-shadow: 0 6rpx 14rpx rgba(245, 158, 11, 0.15);
}

.announcement-label {
	flex-shrink: 0;
	font-size: 22rpx;
	font-weight: 700;
	color: #9a3412;
	padding: 6rpx 12rpx;
	border-radius: 999rpx;
	background: rgba(255, 255, 255, 0.62);
}

.announcement-marquee {
	flex: 1;
	overflow: hidden;
	height: 40rpx;
	display: flex;
	align-items: center;
}

.announcement-track {
	display: inline-flex;
	align-items: center;
	white-space: nowrap;
	animation: announcementMarquee 20s linear infinite;
}

.announcement-track.is-empty {
	animation: none;
}

.announcement-text {
	font-size: 24rpx;
	color: #78350f;
	padding-right: 96rpx;
}

.announcement-more {
	flex-shrink: 0;
	font-size: 22rpx;
	color: #92400e;
	font-weight: 600;
}

@keyframes announcementMarquee {
	0% {
		transform: translateX(0);
	}
	100% {
		transform: translateX(-50%);
	}
}

.header-left {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.menu-btn {
	width: 72rpx;
	height: 72rpx;
	border-radius: 20rpx;
	background-color: #f3f4f6;
	display: flex;
	align-items: center;
	justify-content: center;
	
	.icon {
		font-size: 32rpx;
		color: $text-secondary;
	}
}

.week-selector {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	text-align: left;
	padding: 8rpx 16rpx;
	border-radius: 12rpx;
	transition: background-color 0.2s;
	
	&:active {
		background-color: rgba(0, 0, 0, 0.05);
	}
}

.week-info {
	display: flex;
	align-items: center;
}

.week-text {
	font-size: 32rpx;
	font-weight: 700;
	color: $text-primary;
}

.icon-arrow {
	font-size: 20rpx;
	color: $text-primary;
	margin-left: 8rpx;
}

.week-hint {
	font-size: 20rpx;
	color: $text-light;
	margin-top: 4rpx;
}

.refresh-btn {
	display: flex;
	align-items: center;
	padding: 12rpx 24rpx;
	border-radius: 32rpx;
	background-color: #fff;
	border: 2rpx solid $border-color;
	box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.05);
	
	.icon {
		font-size: 24rpx;
		margin-right: 8rpx;
	}
	
	.refresh-text {
		font-size: 20rpx;
		color: $text-secondary;
	}
}

.date-header {
	display: flex;
}

.time-col {
	width: 64rpx;
	flex-shrink: 0;
}

.day-col {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
}

.day-name {
	font-size: 22rpx;
	font-weight: 700;
	color: $text-primary;
}

.day-date {
	font-size: 18rpx;
	color: $text-light;
	margin-top: 4rpx;
}

.schedule-grid {
	flex: 1;
	background-color: #fff;
	overflow: hidden;
	margin-bottom: 120rpx;
}

.grid-container {
	display: flex;
	padding: 8rpx 16rpx;
	height: 100%;
	position: relative;
}

.time-sidebar {
	width: 64rpx;
	flex-shrink: 0;
	display: flex;
	flex-direction: column;
}

.time-slot {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	border-right: 2rpx dashed $border-color;
	min-height: 0;
	
	&.break-slot {
		flex: 0.6;
		background-color: rgba(251, 191, 36, 0.1);
		border-radius: 8rpx;
		margin: 2rpx 0;
	}
}

.slot-num {
	font-size: 20rpx;
	font-weight: 600;
	color: $text-secondary;
}

.slot-time {
	font-size: 14rpx;
	color: $text-light;
	text-align: center;
	white-space: pre-line;
	line-height: 1.1;
}

.break-label {
	font-size: 18rpx;
	font-weight: 700;
	color: #f59e0b;
}

.courses-area {
	flex: 1;
	position: relative;
}

.break-row {
	position: absolute;
	left: 0;
	right: 0;
	height: calc(0.6 / 11.2 * 100%);
	background-color: rgba(251, 191, 36, 0.05);
	border-radius: 8rpx;
	border: 2rpx solid rgba(251, 191, 36, 0.2);
	display: flex;
	align-items: center;
	justify-content: center;
}

.break-content {
	display: flex;
	align-items: center;
	
	.icon {
		font-size: 24rpx;
		margin-right: 12rpx;
	}
	
	.break-text {
		font-size: 20rpx;
		color: #f59e0b;
		font-weight: 500;
	}
}

.course-card {
	position: absolute;
	padding: 6rpx 8rpx;
	border-radius: 8rpx;
	border-left: 4rpx solid;
	box-shadow: 0 2rpx 6rpx rgba(0,0,0,0.05);
	display: flex;
	flex-direction: column;
	overflow: hidden;
	
	&.color-sky {
		background-color: rgba(14, 165, 233, 0.25);
		border-left-color: #0ea5e9;
	}
	
	&.color-pink {
		background-color: rgba(236, 72, 153, 0.25);
		border-left-color: #ec4899;
	}
	
	&.color-amber {
		background-color: rgba(245, 158, 11, 0.25);
		border-left-color: #f59e0b;
	}
	
	&.color-indigo {
		background-color: rgba(99, 102, 241, 0.25);
		border-left-color: #6366f1;
	}
	
	&.color-yellow {
		background-color: rgba(234, 179, 8, 0.25);
		border-left-color: #eab308;
	}
	
	&.color-emerald {
		background-color: rgba(16, 185, 129, 0.25);
		border-left-color: #10b981;
	}
	
	&.color-purple {
		background-color: rgba(139, 92, 246, 0.25);
		border-left-color: #8b5cf6;
	}
	
	&.color-teal {
		background-color: rgba(20, 184, 166, 0.25);
		border-left-color: #14b8a6;
	}
}

.course-name {
	font-size: 18rpx;
	font-weight: 700;
	color: $text-primary;
	line-height: 1.2;
	word-break: break-all;
	margin-bottom: 2rpx;
}

.course-teacher,
.course-room {
	font-size: 16rpx;
	color: $text-secondary;
	line-height: 1.3;
}

// ============ 公告弹窗样式 ============
.announcement-overlay {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(0, 0, 0, 0.55);
	z-index: 999;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 60rpx 48rpx;
}

.announcement-popup {
	position: relative;
	width: 100%;
	max-height: 70vh;
	background: #fff;
	border-radius: 32rpx;
	padding: 56rpx 40rpx 40rpx;
	display: flex;
	flex-direction: column;
	box-shadow: 0 24rpx 80rpx rgba(0, 0, 0, 0.18);
	animation: popupSlideUp 0.3s ease-out;
}

@keyframes popupSlideUp {
	from {
		opacity: 0;
		transform: translateY(60rpx);
	}
	to {
		opacity: 1;
		transform: translateY(0);
	}
}

.popup-close {
	position: absolute;
	top: 20rpx;
	right: 24rpx;
	width: 56rpx;
	height: 56rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50%;
	background: #f1f5f9;

	.close-icon {
		font-size: 28rpx;
		color: #94a3b8;
		font-weight: 600;
	}
}

.popup-type-badge {
	align-self: flex-start;
	padding: 6rpx 20rpx;
	border-radius: 20rpx;
	margin-bottom: 20rpx;

	.badge-text {
		font-size: 22rpx;
		font-weight: 600;
	}

	&.type-NORMAL {
		background-color: rgba(59, 130, 246, 0.12);
		.badge-text { color: #3b82f6; }
	}
	&.type-IMPORTANT {
		background-color: rgba(245, 158, 11, 0.12);
		.badge-text { color: #f59e0b; }
	}
	&.type-URGENT {
		background-color: rgba(239, 68, 68, 0.12);
		.badge-text { color: #ef4444; }
	}
	&.type-MAINTENANCE {
		background-color: rgba(107, 114, 128, 0.12);
		.badge-text { color: #6b7280; }
	}
}

.popup-title {
	font-size: 34rpx;
	font-weight: 700;
	color: #1e293b;
	line-height: 1.4;
	margin-bottom: 12rpx;
	white-space: normal;
	word-break: break-all;
	overflow: hidden;
}

.popup-time {
	font-size: 24rpx;
	color: #94a3b8;
	margin-bottom: 24rpx;
}

.popup-content {
	flex: 1;
	max-height: 360rpx;
	margin-bottom: 32rpx;
}

.popup-content-text {
	font-size: 28rpx;
	color: #475569;
	line-height: 1.7;
	white-space: pre-wrap;
	word-break: break-all;
}

.popup-image-list {
	display: flex;
	flex-direction: column;
	gap: 16rpx;
	margin-top: 16rpx;
}

.popup-content-image {
	width: 100%;
	border-radius: 12rpx;
	background: #f8fafc;
}

.popup-footer {
	display: flex;
	align-items: center;
	justify-content: flex-end;
	gap: 16rpx;
}

.popup-counter {
	.counter-text {
		font-size: 24rpx;
		color: #94a3b8;
	}
}

.popup-btn {
	background: linear-gradient(135deg, #3b82f6, #2563eb);
	padding: 18rpx 48rpx;
	border-radius: 24rpx;
	
	.btn-text {
		font-size: 28rpx;
		font-weight: 600;
		color: #fff;
	}
}
</style>
