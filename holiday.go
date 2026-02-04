package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HolidayItem 单日节假日信息（API 返回的某一天）
type HolidayItem struct {
	Holiday   bool   `json:"holiday"`    // true=休息日，false=调休工作日
	Name      string `json:"name"`
	Wage      int    `json:"wage"`
	Date      string `json:"date"`      // 如 "2026-01-01"
	CnLunar   string `json:"cnLunar"`
	ExtraInfo string `json:"extra_info"`
	Rest      int    `json:"rest"`
}

// HolidayResponse API 响应，key 为 "MM-DD"
type HolidayResponse struct {
	Code    int                  `json:"code"`
	Holiday map[string]HolidayItem `json:"holiday"`
}

// TodayWorkInfo 今日上班与最近假期信息
type TodayWorkInfo struct {
	IsWorkDay      bool   // 今天是否上班（true=上班，false=休息）
	NearestName    string // 离今天最近的（下一个）假期名称
	DaysToHoliday  int    // 距离该假期的天数
	WeekdayName    string // 今天是礼拜几，如 "周三"
	DaysToWeekend  int    // 距离周末还有几天（周六、周日为 0）
}

const holidayAPIBase = "https://holiday.ailcc.com/api/holiday/year"

var weekdayNames = []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}

// getWeekdayAndDaysToWeekend 返回礼拜几与距离周末的天数（周六、周日为 0）
func getWeekdayAndDaysToWeekend(t time.Time) (weekdayName string, daysToWeekend int) {
	wd := t.Weekday()
	weekdayName = weekdayNames[wd]
	if wd == time.Saturday || wd == time.Sunday {
		daysToWeekend = 0
	} else {
		daysToWeekend = int(time.Saturday - wd)
	}
	return weekdayName, daysToWeekend
}

// FetchHolidayYear 获取指定年份的节假日数据（GET）
func FetchHolidayYear(year int) (*HolidayResponse, error) {
	url := fmt.Sprintf("%s/%d", holidayAPIBase, year)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data HolidayResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.Code != 0 {
		return nil, fmt.Errorf("api code not 0: %d", data.Code)
	}
	return &data, nil
}

// GetTodayWorkInfo 获取今天是否上班、离今天最近的假期名称及还有多少天
func GetTodayWorkInfo() (*TodayWorkInfo, error) {
	now := time.Now()
	year := now.Year()
	todayStr := now.Format("01-02") // MM-DD

	data, err := FetchHolidayYear(year)
	if err != nil {
		return nil, err
	}

	info := &TodayWorkInfo{}

	// 今天是礼拜几、距离周末还有几天
	info.WeekdayName, info.DaysToWeekend = getWeekdayAndDaysToWeekend(now)

	// 今天是否上班：若在 holiday 中且 holiday==true 则为休息日，否则上班
	if item, ok := data.Holiday[todayStr]; ok && item.Holiday {
		info.IsWorkDay = false
	} else {
		info.IsWorkDay = true
	}

	// 找离今天最近的“下一个”假期（含今天算 0 天）
	layout := "2006-01-02"
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var nearestDate time.Time
	for _, v := range data.Holiday {
		if !v.Holiday {
			continue
		}
		t, err := time.Parse(layout, v.Date)
		if err != nil {
			continue
		}
		t = t.In(now.Location())
		if t.Before(todayStart) {
			continue
		}
		if nearestDate.IsZero() || t.Before(nearestDate) {
			nearestDate = t
			info.NearestName = v.Name
		}
	}

	// 若今年没有后续假期，尝试明年的第一个假期
	if info.NearestName == "" {
		nextData, err := FetchHolidayYear(year + 1)
		if err == nil {
			for _, v := range nextData.Holiday {
				if !v.Holiday {
					continue
				}
				t, err := time.Parse(layout, v.Date)
				if err != nil {
					continue
				}
				t = t.In(now.Location())
				if nearestDate.IsZero() || t.Before(nearestDate) {
					nearestDate = t
					info.NearestName = v.Name
				}
			}
		}
	}

	if !nearestDate.IsZero() {
		days := int(nearestDate.Truncate(24*time.Hour).Sub(todayStart).Hours() / 24)
		if days < 0 {
			days = 0
		}
		info.DaysToHoliday = days
	}

	return info, nil
}



