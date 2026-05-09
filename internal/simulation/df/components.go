package df

// Tipos iniciales para refactor Dwarf-Fortress-like

// Item representa un tipo de objeto en el mundo.
type Item struct {
	ID    string
	Name  string
	Stack int // cantidad por stack
}

// Inventory simple asociado a una entidad (NPC, edificio)
type Inventory struct {
	OwnerEntity int64
	Items       map[string]int // itemID -> cantidad
	Cap         int            // capacidad total (por ejemplo, número de stacks)
}

// Job representa una tarea asignable (ej. "mend_sword", "haul_wood", "farm").
type Job struct {
	ID          string
	Role        string // profesión requerida
	TargetX     int
	TargetY     int
	TargetEntity int64
	Priority    int
	Payload     map[string]interface{}
	// Consumes items when starting the job (item -> quantity)
	Consumes    map[string]int
	// Produces items when job completes (item -> quantity)
	Produces    map[string]int
	ActionID    string  // action id to run (e.g., "work_inside")
	Reward      float64 // reward to give on completion
}

// JobQueue es una estructura simple para almacenar jobs por edificio/manager.
type JobQueue struct {
	Jobs []Job
}
