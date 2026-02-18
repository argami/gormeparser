package models

// Seccion constants
const (
	SeccionDefault = SeccionA
)

// Subseccion constants
type Subseccion string

const (
	SubseccionActosInscritos Subseccion = "A"
	SubseccionOtrosActos     Subseccion = "B"
)

// ActoCargo represents cargo types that have arguments (appointments, cessations, etc.)
type ActoCargo string

const (
	ActoCargoNombramientos       ActoCargo = "Nombramientos"
	ActoCargoRevocaciones        ActoCargo = "Revocaciones"
	ActoCargoCesesDimisiones     ActoCargo = "Ceses/Dimisiones"
	ActoCargoConstitucion        ActoCargo = "Constitucion"
	ActoCargoDisolucion         ActoCargo = "Disolucion"
	ActoCargoFinCuadro          ActoCargo = "FinCuadro"
	ActoCargoReeleccion         ActoCargo = "Reeleccion"
	ActoCargoNombramiento        ActoCargo = "Nombramiento"
	ActoCargoOtroCargo          ActoCargo = "OtroCargo"
)

// ActoNoArg represents act types without arguments
type ActoNoArg string

const (
	ActoNoArgCreditoIncobrable     ActoNoArg = "CreditoIncobrable"
	ActoNoArgSociedadUnipersonal   ActoNoArg = "SociedadUnipersonal"
	ActoNoArgExtincion            ActoNoArg = "Extincion"
	ActoNoArgCuadroCargos         ActoNoArg = "CuadroCargos"
	ActoNoArgCambioObjetoSocial    ActoNoArg = "CambioObjetoSocial"
	ActoNoArgOtro                 ActoNoArg = "Otro"
)

// ActoColon represents act types with colon arguments
type ActoColon string

const (
	ActoColonModificacionDuracion ActoColon = "Modificacion de duracion"
	ActoColonFeDeErratas         ActoColon = "Fe de erratas"
	ActoColonDomicilio           ActoColon = "Domicilio"
	ActoColonObjeto              ActoColon = "Objeto"
	ActoColonCapital             ActoColon = "Capital"
	ActoColonEstatutos           ActoColon = "Estatutos"
	ActoColonDenominacion        ActoColon = "Denominacion"
)

// ActoBold represents bold act types
type ActoBold string

const (
	ActoBoldUnipersonalidad   ActoBold = "Declaracion de unipersonalidad"
	ActoBoldSociedadUnipersonal ActoBold = "Sociedad unipersonal"
	ActoBoldEscisionTotal     ActoBold = "Escision total"
	ActoBoldFusion            ActoBold = "Fusion"
	ActoBoldDisolucion        ActoBold = "Disolucion"
)
