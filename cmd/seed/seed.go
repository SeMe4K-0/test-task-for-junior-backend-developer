package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"example.com/taskservice/internal/config"
	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	"example.com/taskservice/internal/domain/task"
	taskdomain "example.com/taskservice/internal/domain/task"
	postgresinfra "example.com/taskservice/internal/infrastructure/postgres"
	postgrestask "example.com/taskservice/internal/repository/postgres"
)

// Gen Tasks

var titles = []string{
	"Провести утренний обход пациентов", "Заполнить истории болезни", "Назначить лабораторные анализы", "Проверить жизненные показатели", "Выписать рецепты на лекарства", "Провести вакцинацию детей", "Обновить карты диспансерного наблюдения", "Подготовить пациентов к операции", "Провести ЭКГ исследование", "Выдать больничные листы", "Проверить срок годности медикаментов", "Провести перевязку послеоперационных ран", "Собрать анамнез у новых пациентов", "Направить на инструментальную диагностику", "Провести санитарную обработку палат", "Проверить результаты МРТ", "Провести реанимационные мероприятия", "Оформить эпикризы по выписке", "Провести забор крови на анализ", "Заполнить журнал учёта процедур",
}

var descriptions = []string{
	"Обойти 15 пациентов в терапевтическом отделении, зафиксировать жалобы", "Внести данные о лечении в электронные истории болезни по стандарту", "Выдать направления на ОАК, биохимию и коагулограмму для 10 пациентов", "Измерить АД, пульс, сатурацию и температуру у всех поступивших", "Выписать антибиотики и противовирусные согласно назначениям врача", "Поставить вакцину от гриппа и гепатита B в прививочном кабинете", "Обновить статусы хронических больных в регистре диспансеризации", "Провести предоперационную подготовку, забор анализов и ЭКГ", "Снять и расшифровать ЭКГ у пациентов кардиологического отделения", "Оформить и выдать листки нетрудоспособности сотрудникам поликлиники", "Провести ревизию аптечки: просроченные препараты утилизировать", "Сделать перевязку 8 пациентам после полостных операций", "Собрать жалобы, аллергоанамнез и сопутствующие заболевания", "Направить на УЗИ, КТ и рентген по клиническим показаниям", "Обработать палаты бактерицидными лампами и дезрастворами", "Расшифровать и подшить в карты результаты магнитно-резонансной томографии", "Провести СЛР при остановке дыхания и кровообращения по алгоритму", "Написать выписные эпикризы для 5 пациентов с инфарктом", "Взять венозную кровь у 12 пациентов на гормоны и онкомаркеры", "Записать все выполненные инъекции и капельницы в процедурный журнал",
}

var statuses = []task.Status{
	"new", "in_progress", "done", "new", "in_progress", "done", "canceled", "new", "in_progress", "done", "new", "in_progress", "done", "canceled", "new", "in_progress", "done", "new", "in_progress", "done",
}

var evenOdd = []recurrencedomain.EvenOdd{
	"even", "odd",
}

func generateTasks(count int) ([]taskdomain.Task, error) {
	var tasks []taskdomain.Task
	for i := 0; i < count; i++ {
		var dueDate *time.Time
		if rand.IntN(10) > 3 {
			date := time.Now().Add(time.Hour * 24 * time.Duration(rand.IntN(30)))
			dueDate = &date
		}
		createdAt := time.Now().Add(time.Hour * 24 * -time.Duration(rand.IntN(10)))
		taskData := taskdomain.Task{
			Title:       titles[rand.IntN(len(titles))],
			Description: descriptions[rand.IntN(len(descriptions))],
			DueDate:     dueDate,
			Status:      statuses[rand.IntN(len(statuses))],
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt.Add(time.Hour * 24 * time.Duration(rand.IntN(10))),
		}
		tasks = append(tasks, taskData)
	}
	return tasks, nil
}

// Gen Recurrence

func fillInterval(recurrence *recurrencedomain.Recurrence) {
	intervalDays := rand.IntN(9) + 1
	recurrence.IntervalDays = &intervalDays
}

func fillMonthDays(recurrence *recurrencedomain.Recurrence) {
	var monthDays []int
	for i := 0; i < (rand.IntN(10)); i++ {
		monthDays = append(monthDays, rand.IntN(30)+1)
	}
	recurrence.MonthDays = &monthDays
}

func fillSpecificDates(recurrence *recurrencedomain.Recurrence) {
	var specificDates []time.Time
	for i := 0; i < (rand.IntN(10)); i++ {
		specificDates = append(specificDates, time.Now().Add(time.Hour*24*time.Duration(rand.IntN(10))))
	}
	recurrence.SpecificDates = &specificDates
}

func fillEvenOdd(recurrence *recurrencedomain.Recurrence) {
	recurrence.EvenOdd = &evenOdd[rand.IntN(2)]
}

func generateRecurrences(count int) ([]recurrencedomain.Recurrence, error) {
	var recurrences []recurrencedomain.Recurrence
	for i := 0; i < count; i++ {
		createdAt := time.Now().Add(time.Hour * 24 * -time.Duration(rand.IntN(10)))
		startDate := time.Now().Add(time.Hour * 24 * -time.Duration(rand.IntN(10)))
		var endDate *time.Time
		if rand.IntN(3) == 0 {
			t := time.Now().Add(time.Hour * 24 * time.Duration(rand.IntN(10)))
			endDate = &t
		}

		recurrenceData := recurrencedomain.Recurrence{
			Title:       titles[rand.IntN(len(titles))],
			Description: descriptions[rand.IntN(len(descriptions))],
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt.Add(time.Hour * 24 * time.Duration(rand.IntN(10))),
			StartDate:   &startDate,
			EndDate:     endDate,
		}
		recurrenceType := rand.IntN(4)

		switch recurrenceType {
		case 0:
			fillInterval(&recurrenceData)
		case 1:
			fillMonthDays(&recurrenceData)
		case 2:
			fillSpecificDates(&recurrenceData)
		case 3:
			fillEvenOdd(&recurrenceData)
		}

		recurrences = append(recurrences, recurrenceData)
	}
	return recurrences, nil
}

func main() {
	dsn := config.LoadConfig().DatabaseDSN
	pool, err := postgresinfra.Open(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	repo := postgrestask.New(pool)

	// fill tasks db
	tasks, err := generateTasks(20)
	if err != nil {
		log.Fatal(err)
	}

	for _, val := range tasks {
		_, err := repo.Create(context.Background(), &val)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Созданы 20 разовых задач")

	// fill recurrences db

	recurrences, err := generateRecurrences(20)
	if err != nil {
		log.Fatal(err)
	}
	for _, val := range recurrences {
		_, err := repo.CreateRecurrence(context.Background(), &val)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Созданы 20 повторяющихся задач")
}
