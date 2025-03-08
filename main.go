package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
)

type TemplateData struct {
	TaskOne     *TaskOneData
	TaskTwo     *TaskTwoData
	TaskThree   *TaskThreeData
	ActiveTab   string
	Task1Result bool
	Task2Result bool
	Task3Result bool
}

type TaskOneData struct {
	IK     float64
	TF     float64
	SM     float64
	TM     float64
	Result string
}

type TaskTwoData struct {
	S      float64
	Result string
}

type TaskThreeData struct {
	RCN    float64
	XCN    float64
	RCMin  float64
	XCMin  float64
	Result string
}

func calculateTaskOne(iK, tF, sM, tM float64) *TaskOneData {
	u := 10.0
	iM := (sM / 2.0) / (math.Sqrt(3.0) * u)

	var jEK float64
	if tM >= 1000 && tM < 3000 {
		jEK = 1.6
	} else if tM >= 3000 && tM <= 5000 {
		jEK = 1.4
	} else {
		jEK = 1.2
	}

	s := iM / jEK
	sMin := (iK * 1000 * math.Sqrt(tF)) / 92.0

	result := ""
	if s >= sMin {
		result = fmt.Sprintf("Переріз жил кабелю: %.2f мм.", s)
	} else {
		result = fmt.Sprintf("Переріз жил кабелю %.2f мм і має бути збільшений мінімум до %.2f мм.", s, sMin)
	}

	return &TaskOneData{
		IK:     iK,
		TF:     tF,
		SM:     sM,
		TM:     tM,
		Result: result,
	}
}

func calculateTaskTwo(s float64) *TaskTwoData {
	uSN := 10.5
	uK := 10.5
	sNom := 6.3

	xC := math.Pow(uSN, 2) / s
	xT := (uK / 100.0) * (math.Pow(uSN, 2) / sNom)
	xSum := xC + xT
	iP0 := uSN / (math.Sqrt(3) * xSum)

	return &TaskTwoData{
		S:      s,
		Result: fmt.Sprintf("Початкове діюче значення струму трифазного КЗ: %.2f кА.", iP0),
	}
}

func calculateTaskThree(rCN, xCN, rCMin, xCMin float64) *TaskThreeData {
	uMax := 11.1
	uVN := 115.0
	sNom := 6.3
	uNN := 11.0

	xT := (uMax * math.Pow(uVN, 2)) / (100 * sNom)
	zSh := math.Sqrt(math.Pow(rCN, 2) + math.Pow(xCN+xT, 2))
	zShMin := math.Sqrt(math.Pow(rCMin, 2) + math.Pow(xCMin+xT, 2))

	i3Sh := (uVN * 1000) / (1.73 * zSh)
	i2Sh := i3Sh * (1.73 / 2)
	i3ShMin := (uVN * 1000) / (1.73 * zShMin)
	i2ShMin := i3ShMin * (1.73 / 2)

	kPR := math.Pow(uNN, 2) / math.Pow(uVN, 2)
	kPR = math.Round(kPR*1000) / 1000.0

	zShN := math.Sqrt(math.Pow(rCN*kPR, 2) + math.Pow((xCN+xT)*kPR, 2))
	zShNMin := math.Sqrt(math.Pow(rCMin*kPR, 2) + math.Pow((xCMin+xT)*kPR, 2))

	i3ShN := (uNN * 1000) / (1.73 * zShN)
	i2ShN := i3ShN * (1.73 / 2)
	i3ShNMin := (uNN * 1000) / (1.73 * zShNMin)
	i2ShNMin := i3ShNMin * (1.73 / 2)

	l := 12.37
	rL := l * 0.64
	xL := l * 0.363

	zEN := math.Sqrt(math.Pow(rL+rCN*kPR, 2) + math.Pow(xL+(xCN+xT)*kPR, 2))
	zENMin := math.Sqrt(math.Pow(rL+rCMin*kPR, 2) + math.Pow(xL+(xCMin+xT)*kPR, 2))

	i3LN := (uNN * 1000) / (1.73 * zEN)
	i2LN := i3LN * (1.73 / 2)
	i3LNMin := (uNN * 1000) / (1.73 * zENMin)
	i2LNMin := i3LNMin * (1.73 / 2)

	result := fmt.Sprintf(`Струми трифазного та двофазного КЗ на шинах 10кВ в нормальному та мінімальному режимах, приведені до напруги 110 кВ:
Опір трифазного кз: нормальний: %.2f A, мінімальний: %.2f A
Опір двофазного кз: нормальний: %.2f A, мінімальний: %.2f A
Дійсні струми трифазного та двофазного КЗ на шинах 10кВ в нормальному та мінімальному режимах:
Опір трифазного кз: нормальний: %.2f A, мінімальний: %.2f A
Опір двофазного кз: нормальний: %.2f A, мінімальний: %.2f A
Струми трифазного та двофазного КЗ в точці 10 в нормальному та мінімальному режимах:
Опір трифазного кз: нормальний: %.2f A, мінімальний: %.2f A
Опір двофазного кз: нормальний: %.2f A, мінімальний: %.2f A
Аварійний режим на цій підстанції не передбачений`,
		i3Sh, i3ShMin, i2Sh, i2ShMin,
		i3ShN, i3ShNMin, i2ShN, i2ShNMin,
		i3LN, i3LNMin, i2LN, i2LNMin)

	return &TaskThreeData{
		RCN:    rCN,
		XCN:    xCN,
		RCMin:  rCMin,
		XCMin:  xCMin,
		Result: result,
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	data := &TemplateData{
		ActiveTab: "task1", // Активна вкладка за замовчуванням
	}

	if r.Method == http.MethodPost {
		r.ParseForm()

		// Оновлення активної вкладки, якщо необхідно
		if tab := r.FormValue("activeTab"); tab != "" {
			data.ActiveTab = tab
		}

		switch r.FormValue("task") {
		case "1":
			iK, _ := strconv.ParseFloat(r.FormValue("ik"), 64)
			tF, _ := strconv.ParseFloat(r.FormValue("tf"), 64)
			sM, _ := strconv.ParseFloat(r.FormValue("sm"), 64)
			tM, _ := strconv.ParseFloat(r.FormValue("tm"), 64)
			data.TaskOne = calculateTaskOne(iK, tF, sM, tM)
			data.Task1Result = true

		case "2":
			s, _ := strconv.ParseFloat(r.FormValue("s"), 64)
			data.TaskTwo = calculateTaskTwo(s)
			data.Task2Result = true

		case "3":
			rCN, _ := strconv.ParseFloat(r.FormValue("rcn"), 64)
			xCN, _ := strconv.ParseFloat(r.FormValue("xcn"), 64)
			rCMin, _ := strconv.ParseFloat(r.FormValue("rcmin"), 64)
			xCMin, _ := strconv.ParseFloat(r.FormValue("xcmin"), 64)
			data.TaskThree = calculateTaskThree(rCN, xCN, rCMin, xCMin)
			data.Task3Result = true
		}
	}

	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
