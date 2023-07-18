package lazy_save

type LazySaveObj interface {
	GetLsoId() string
	SaveOrUpdate(callback func())
}
