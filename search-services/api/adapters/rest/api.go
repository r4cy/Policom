package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"yadro.com/course/api/core"
)

type PingReply struct {
	Replies map[string]string `json:"replies"`
}

// @Summary      Проверка сервисов.
// @Description  Опрашивает все зарегистрированные микросервисы и возвращает их статус. Используется для health-check мониторинга.
// @Tags         ping
// @Produce      json
// @Success      200 {object} PingReply
// @Failure		 400 "Неверные параметры запроса"
// @Failure		 500 "Внутренняя ошибка сервера"
// @Router       /api/ping [get]
func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		array := make(map[string]string)
		if pingers == nil {
			http.Error(w, "empty argument for pingers", http.StatusBadRequest)
			return
		}

		for name, ping := range pingers {
			err := ping.Ping(r.Context())
			if err != nil {
				array[name] = "unavailable"
				continue
			}
			array[name] = "ok"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(PingReply{Replies: array}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewPingHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

// @Summary      Обновление комиксов в БД и Индексе.
// @Description  Загружает новые комиксы с XKCD и обновляет поисковый индекс. Если обновление уже запущено - возвращает 202.
// @Tags         update
// @Produce      json
// @Success      200 "Обновление запущено"
// @Success      202 "Обновление БД"
// @Failure		 500 "Ошибка, сервер в данный момент обновляет БД"
// @Router       /api/update [post]
func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			if errors.Is(err, core.ErrAlreadyRunning) {
				log.ErrorContext(r.Context(),
					"failed to do update",
					"handler", "NewUpdateHandler",
					"err", core.ErrAlreadyRunning,
				)
				w.WriteHeader(http.StatusAccepted)
				return
			}
			http.Error(w, "failed to do update because server is running", http.StatusInternalServerError)
		}
	}
}

type StatsReply struct {
	WordsTotal    int `json:"words_total" example:"10000" minimum:"0"`
	WordsUnique   int `json:"words_unique" example:"2555" minimum:"0"`
	ComicsFetched int `json:"comics_fetched" example:"3190" minimum:"0"`
	ComicsTotal   int `json:"comics_total" example:"3200" minimum:"0"`
}

// @Summary      Статистика из базы данных.
// @Description  Возвращает агрегированную статистику по комиксам в БД: общее кол-во слов, уникальных слов, комиксов загруженных при последнем обновлении и всего в базе.
// @Tags         stats
// @Produce      json
// @Success      200 {object} StatsReply
// @Failure		 500 "Ошибка на стороне сервера"
// @Router       /api/db/stats [get]
func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := updater.Stats(r.Context())
		if err != nil {
			http.Error(w, "failed to make stats about comics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(StatsReply{
			WordsTotal:    resp.WordsTotal,
			WordsUnique:   resp.WordsUnique,
			ComicsFetched: resp.ComicsFetched,
			ComicsTotal:   resp.ComicsTotal,
		}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewUpdateStatsHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

type UpdateStatusReply struct {
	Status core.UpdateStatus `json:"status" example:"running"`
}

// @Summary      Статус процесса обновления.
// @Description  Возвращает текущий статус процесса обновления БД: idle или running.
// @Tags         status
// @Produce      json
// @Success      200 {object} UpdateStatusReply
// @Failure		 500 "Ошибка на стороне сервера"
// @Router       /api/db/status [get]
func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := updater.Status(r.Context())
		if err != nil {
			http.Error(w, "failed to take status about server", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(UpdateStatusReply{Status: resp}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewUpdateStatusHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

// @Summary      Удаление данных из БД.
// @Description  Удаляет все данные из таблицы. ВНИМАНИЕ: все данные будут безвозвратно удалены.
// @Tags         database
// @Success      200 "Успешное удаление данных"
// @Failure		 500 "Ошибка на стороне сервера"
// @Router       /api/db [delete]
func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := updater.Drop(r.Context())
		if err != nil {
			http.Error(w, "failed to drop tables", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

type SearchReply struct {
	Comics []core.Comics `json:"comics"`
	Total  int           `json:"total" example:"0" minimum:"0"`
}

// @Summary      Поиск комиксов по ключевым словам.
// @Description  Ищет комиксы в базе данных по переданной фразе. Результаты отсортированы по релевантности.
// @Tags         search
// @Produce      json
// @Param        phrase  query  string  true   "Поисковая фраза"
// @Param  		 limit  query  int  false  "Макс. кол-во результатов"  minimum(0)
// @Success      200 {object} SearchReply
// @Failure		 400 "Неверные параметры запроса"
// @Failure		 500 "Ошибка на стороне сервера"
// @Router       /api/search [get]
func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")

		limitStr := r.URL.Query().Get("limit")
		limit := 0
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit < 0 {
				http.Error(w, "bad limit in request", http.StatusBadRequest)
				return
			}
		}

		comics, err := searcher.Search(r.Context(), phrase, limit)
		if err != nil {
			http.Error(w, "search server takes error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(SearchReply{
			Comics: comics,
			Total:  len(comics),
		}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewSearchHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

// @Summary      Поиск комиксов по индексу (in-memory)
// @Description  Ищет комиксы в in-memory индексе по переданной фразе. Быстрее чем /api/search, результаты из памяти.
// @Tags         search
// @Produce      json
// @Param        phrase  query  string  true   "Поисковая фраза"
// @Param  		 limit  query  int  false  "Макс. кол-во результатов"  minimum(0)
// @Success      200 {object} SearchReply
// @Failure		 400 "Неверные параметры запроса"
// @Failure		 500 "Ошибка на стороне сервера"
// @Router       /api/isearch [get]
func NewISearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")

		limitStr := r.URL.Query().Get("limit")
		limit := 0
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit < 0 {
				http.Error(w, "bad limit in request", http.StatusBadRequest)
				return
			}
		}

		comics, err := searcher.ISearch(r.Context(), phrase, limit)
		if err != nil {
			http.Error(w, "search server takes error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(SearchReply{
			Comics: comics,
			Total:  len(comics),
		}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewISearchHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginReply struct {
	Token string `json:"token"`
}

// @Summary      Авторизация пользователя.
// @Description  Принимает имя и пароль, возвращает JWT токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  LoginRequest  true  "Данные для входа"
// @Success  	 200  {object}  LoginReply  "JWT токен"
// @Failure		 400 "Неверные параметры запроса"
// @Failure		 401 "Неверный логин или пароль"
// @Failure		 500 "Ошибка на стороне сервера"
// @Router       /api/login [post]
func NewLoginHandler(log *slog.Logger, login core.AAA) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &LoginRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON request",
				"handler", "NewLoginHandler",
				"err", err,
			)
			http.Error(w, "bad name or password in request", http.StatusBadRequest)
			return
		}

		token, err := login.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			log.ErrorContext(r.Context(),
				"failed to authorized",
				"handler", "NewLoginHandler",
				"err", err,
			)
			http.Error(w, "failed to authorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(LoginReply{
			Token: token,
		}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewLoginHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password1 string `json:"password1"`
	Password2 string `json:"password2"`
}

// @Summary      Регистрация пользователя
// @Description  Создает нового пользователя с ролью user. Пароли password1 и password2 должны совпадать.
// @Tags         auth
// @Accept       json
// @Produce      plain
// @Param        body  body  RegisterRequest  true  "Данные для регистрации"
// @Success      201  "Пользователь успешно создан"
// @Failure      400  "Некорректный JSON или пароли не совпадают"
// @Failure      500  "Ошибка регистрации пользователя"
// @Router       /api/auth/register [post]
func NewRegisterHandler(log *slog.Logger, login core.AAA) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &RegisterRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON request",
				"handler", "NewRegisterHandler",
				"err", err,
			)
			http.Error(w, "bad parametres in request", http.StatusBadRequest)
			return
		}

		if req.Password1 != req.Password2 {
			http.Error(w, "passwords not equal in request", http.StatusBadRequest)
			return
		}

		err = login.Register(r.Context(), req.Username, req.Email, req.Password1, core.RoleUser)
		if err != nil {
			log.ErrorContext(r.Context(),
				"failed to register",
				"handler", "NewRegisterHandler",
				"err", err,
			)
			http.Error(w, "failed to authorized", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// @Summary      Получение комикса по ID
// @Description  Возвращает данные одного комикса по его идентификатору.
// @Tags         comics
// @Produce      json
// @Param        id   path  int  true  "ID комикса"  minimum(1)
// @Success      200  {object}  core.Comics  "Комикс найден"
// @Failure      400  "Некорректный ID"
// @Failure      404  "Комикс не найден"
// @Failure      500  "Ошибка сервиса поиска"
// @Router       /api/comics/{id} [get]
func NewComicsHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		var id int

		if idStr == "" {
			http.Error(w, "bad id in request", http.StatusBadRequest)
			return
		} else {
			var err error
			id, err = strconv.Atoi(idStr)
			if err != nil || id < 0 {
				http.Error(w, "bad id in request", http.StatusBadRequest)
				return
			}
		}

		comics, err := searcher.GetByID(r.Context(), id)
		if err != nil {
			if err == core.ErrNotFound {
				http.Error(w, "comics not found", http.StatusNotFound)
				return
			}
			http.Error(w, "search server takes error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(core.Comics{
			ID:             comics.ID,
			Title:          comics.Title,
			URL:            comics.URL,
			Description:    comics.Description,
			ImgDescription: comics.ImgDescription,
		}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewComicsHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

type ProfileReply struct {
	SearchReply
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// @Summary      Профиль пользователя
// @Description  Возвращает данные текущего авторизованного пользователя и список сохраненных комиксов.
// @Tags         profile
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ProfileReply  "Профиль пользователя"
// @Failure      401  "Пользователь не авторизован"
// @Failure      500  "Ошибка получения профиля"
// @Router       /api/me [get]
func NewProfileHandler(log *slog.Logger, profile core.Profile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(core.UserKey).(core.User)
		if !ok {
			http.Error(w, "unauthorized user", http.StatusUnauthorized)
			return
		}

		comics, err := profile.LikesComics(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "search AAA takes error ", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ProfileReply{
			SearchReply: SearchReply{
				Comics: comics,
				Total:  len(comics),
			},
			Username: user.Username,
			Email:    user.Email,
			Role:     string(user.Role),
		}); err != nil {
			log.ErrorContext(r.Context(),
				"failed to make JSON response",
				"handler", "NewComicsSavedHandler",
				"err", err,
			)
			http.Error(w, "failed to make JSON response", http.StatusInternalServerError)
			return
		}
	}
}

// @Summary      Сохранение комикса
// @Description  Добавляет комикс в избранное текущего авторизованного пользователя. Повторное сохранение того же комикса не создает дубликат.
// @Tags         profile
// @Security     BearerAuth
// @Param        id   path  int  true  "ID комикса"  minimum(1)
// @Success      204  "Комикс сохранен"
// @Failure      400  "Некорректный ID"
// @Failure      401  "Пользователь не авторизован"
// @Failure      500  "Ошибка сохранения комикса"
// @Router       /api/me/saved/{id} [post]
func NewSaveComicsHandler(log *slog.Logger, profile core.Profile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(core.UserKey).(core.User)
		if !ok {
			http.Error(w, "unauthorized user", http.StatusUnauthorized)
			return
		}

		idStr := r.PathValue("id")
		var comics_id int

		if idStr == "" {
			http.Error(w, "bad id in request", http.StatusBadRequest)
			return
		} else {
			var err error
			comics_id, err = strconv.Atoi(idStr)
			if err != nil || comics_id < 0 {
				http.Error(w, "bad id in request", http.StatusBadRequest)
				return
			}
		}

		err := profile.LikeComics(r.Context(), user.ID, comics_id)
		if err != nil {
			http.Error(w, "search AAA takes error ", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary      Удаление комикса из избранного
// @Description  Удаляет комикс из избранного текущего авторизованного пользователя. Если комикс уже не был сохранен, операция все равно завершается успешно.
// @Tags         profile
// @Security     BearerAuth
// @Param        id   path  int  true  "ID комикса"  minimum(1)
// @Success      204  "Комикс удален из избранного"
// @Failure      400  "Некорректный ID"
// @Failure      401  "Пользователь не авторизован"
// @Failure      500  "Ошибка удаления комикса из избранного"
// @Router       /api/me/saved/{id} [delete]
func NewUnsaveComicsHandler(log *slog.Logger, profile core.Profile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(core.UserKey).(core.User)
		if !ok {
			http.Error(w, "unauthorized user", http.StatusUnauthorized)
			return
		}

		idStr := r.PathValue("id")
		var comics_id int

		if idStr == "" {
			http.Error(w, "bad id in request", http.StatusBadRequest)
			return
		} else {
			var err error
			comics_id, err = strconv.Atoi(idStr)
			if err != nil || comics_id < 0 {
				http.Error(w, "bad id in request", http.StatusBadRequest)
				return
			}
		}

		err := profile.DiselikeComics(r.Context(), user.ID, comics_id)
		if err != nil {
			http.Error(w, "search AAA takes error ", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
