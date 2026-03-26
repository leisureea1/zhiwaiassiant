/**
 * AI 服务 - 调用 Gemini API
 */

const API_KEY = ''; // 在这里配置你的 API Key

/**
 * 生成AI响应
 * @param prompt 用户输入
 * @returns AI响应文本
 */
export const generateAIResponse = async (prompt: string): Promise<string> => {
  if (!API_KEY) {
    return mockAIResponse(prompt);
  }

  try {
    const response = await uni.request({
      url: 'https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent',
      method: 'POST',
      header: {
        'Content-Type': 'application/json',
        'x-goog-api-key': API_KEY
      },
      data: {
        contents: [{
          parts: [{ text: prompt }]
        }],
        systemInstruction: {
          parts: [{ 
            text: "You are Zhiwai Assistant, a helpful and friendly university companion for students. You help with schedules, campus info, and studying tips. Keep answers concise and respond in Chinese when the user uses Chinese." 
          }]
        }
      }
    });

    const data = response.data as any;
    if (data?.candidates?.[0]?.content?.parts?.[0]?.text) {
      return data.candidates[0].content.parts[0].text;
    }
    return "抱歉，我无法生成回复。";
  } catch (error) {
    console.error("Gemini API Error:", error);
    return "网络连接出现问题，请稍后再试。";
  }
};

/**
 * 模拟AI响应（当没有API Key时使用）
 */
const mockAIResponse = (input: string): string => {
  const lowerInput = input.toLowerCase();
  
  if (lowerInput.includes('课') || lowerInput.includes('schedule')) {
    return '你今天有3节课：\n• 上午 实践俄语III（8:00-9:50，JB107）\n• 下午 国际法学（14:00-15:50，JB107）\n• 晚上 毛泽东思想（19:10-21:00，JA410）';
  }
  
  if (lowerInput.includes('考试') || lowerInput.includes('exam')) {
    return '你最近的考试是1月10日的概率论与数理统计，在教二JA101，时间是9:00-11:00。还有3天时间准备，加油！💪';
  }
  
  if (lowerInput.includes('成绩') || lowerInput.includes('grade') || lowerInput.includes('gpa')) {
    return '你本学期的GPA是3.86，平均分91.5分。\n\n最高分课程：C语言程序设计（98分）\n\n继续保持优秀的学习态度！🎉';
  }
  
  if (lowerInput.includes('校车') || lowerInput.includes('bus')) {
    return '最近的校车班次是14:30发车，预计15:15到达，还有12分钟。目前还有18个座位可用。';
  }
  
  if (lowerInput.includes('天气') || lowerInput.includes('weather')) {
    return '今天校园天气晴朗，温度18-25°C，非常适合户外活动。记得多喝水哦！☀️';
  }
  
  if (lowerInput.includes('图书馆') || lowerInput.includes('library')) {
    return '图书馆开放时间：\n• 周一至周五：8:00-22:00\n• 周末：9:00-21:00\n\n目前一楼阅览室有较多空位可用。';
  }
  
  if (lowerInput.includes('食堂') || lowerInput.includes('饭') || lowerInput.includes('吃')) {
    return '校内有3个食堂：\n• 一食堂（6:30-20:30）\n• 二食堂（7:00-21:00）\n• 清真食堂（7:00-20:00）\n\n推荐今天的特色菜：红烧排骨！';
  }
  
  const responses = [
    '好的，让我帮你查一下相关信息。请问你具体想了解什么呢？',
    '这是一个很好的问题！你可以告诉我更多细节，我会尽力帮助你。',
    '我理解你的问题。你可以尝试询问课程、考试、成绩、校车等相关信息。',
    '感谢你的提问！作为你的校园助手，我可以帮你查询课表、考试安排、成绩等信息。',
    '你好！有什么我可以帮助你的吗？试着问我关于校园生活的问题吧！'
  ];
  
  return responses[Math.floor(Math.random() * responses.length)];
};
