package service

type SyncObject interface {
	ParseResponse(jsonResponse string)
}

type BaseSyncObject struct{}
