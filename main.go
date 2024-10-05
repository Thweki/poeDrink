package main

import (
	"strconv"
	"sync"
	"time"
	"image/color"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2"
	"github.com/go-vgo/robotgo"
)

var (
	isDrinking      bool
	mutex           sync.Mutex
	statusLamp      *canvas.Rectangle
	drinkInterval   = 1500 // 默认喝药间隙为1.5秒
	detectionInterval = 100 // 默认检测间隙为100毫秒
	thresholdOptions = []string{"70", "50", "30"} // 喝药阈值选项
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("喝药控制")

	// 设置窗口的初始大小
	myWindow.Resize(fyne.NewSize(400, 300))

	// 喝药阈值选择
	percentEntry := widget.NewSelect(thresholdOptions, func(selected string) {
		// 选择阈值时的回调
	})

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("输入按键 (例如 '1')")

	frequencyEntry := widget.NewEntry()
	frequencyEntry.SetPlaceHolder("输入检测频率 (毫秒)")

	intervalEntry := widget.NewEntry()
	intervalEntry.SetPlaceHolder("输入喝药间隙 (毫秒)")
	intervalEntry.SetText("1500") // 默认值1.5秒

	// 状态指示灯
	statusLamp = canvas.NewRectangle(color.Gray{Y: 128}) // 初始为灰色
	statusLamp.SetMinSize(fyne.NewSize(20, 20))

	startButton := widget.NewButton("开始喝药", func() {
		key := keyEntry.Text
		frequencyStr := frequencyEntry.Text
		intervalStr := intervalEntry.Text

		frequency, err := strconv.Atoi(frequencyStr)
		if err != nil || frequency <= 0 {
			dialog.ShowInformation("错误", "请输入有效的检测频率 (大于 0)。", myWindow)
			return
		}

		drinkInterval, err = strconv.Atoi(intervalStr)
		if err != nil || drinkInterval <= 0 {
			dialog.ShowInformation("错误", "请输入有效的喝药间隙 (大于 0)。", myWindow)
			return
		}

		mutex.Lock()
		isDrinking = true
		mutex.Unlock()
		updateStatusLamp() // 更新状态指示灯为绿色

		go func() {
			for {
				mutex.Lock()
				if !isDrinking {
					mutex.Unlock()
					break
				}
				mutex.Unlock()

				drink(key, percentEntry.Selected)
				time.Sleep(time.Duration(frequency) * time.Millisecond) // 使用用户设置的频率
			}
		}()
	})

	stopButton := widget.NewButton("停止喝药", func() {
		mutex.Lock()
		isDrinking = false
		mutex.Unlock()
		updateStatusLamp() // 更新状态指示灯为灰色
		dialog.ShowInformation("提示", "已停止喝药。", myWindow)
	})

	content := container.NewVBox(
		widget.NewLabel("设置喝药参数"),
		widget.NewLabel("喝药阈值:"),
		percentEntry,
		widget.NewLabel("按键:"),
		keyEntry,
		widget.NewLabel("检测频率 (毫秒):"),
		frequencyEntry,
		widget.NewLabel("喝药间隙 (毫秒):"),
		intervalEntry,
		startButton,
		stopButton,
		statusLamp,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func updateStatusLamp() {
	if isDrinking {
		statusLamp.FillColor = color.RGBA{0, 255, 0, 255} // 绿色
	} else {
		statusLamp.FillColor = color.Gray{Y: 128} // 灰色
	}
	statusLamp.Refresh() // 刷新状态指示灯
}

func drink(key string, percent string) {
	needDrink := false
	switch percent {
	case "70":
		needDrink = robotgo.GetPixelColor(60, 932) != "3d1e1b"
	case "50":
		needDrink = robotgo.GetPixelColor(56, 976) != "531d17"
	case "30":
		needDrink = robotgo.GetPixelColor(65, 1012) != "460e12"
	}

	if needDrink && robotgo.GetPixelColor(1892, 967) == "9b8974" {
		robotgo.KeyTap(key)
		time.Sleep(time.Duration(drinkInterval) * time.Millisecond) // 喝药间隙
	}
}
