package container

func (a *App) Run() error {
	return a.server.ListenAndServe()
}
