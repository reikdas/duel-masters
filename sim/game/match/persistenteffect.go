package match

type PersistentHandlerFunc func(ctx2 *Context, exit func())

type PersistentEffect struct {
	exit   func()
	effect PersistentHandlerFunc
}

func (match *Match) ApplyPersistentEffect(f PersistentHandlerFunc) {
	match.persistentSeq++
	key := match.persistentSeq

	fx := PersistentEffect{
		exit:   func() { match.RemovePersistentEffect(key) },
		effect: f,
	}

	match.persistentEffects[key] = fx
}

func (match *Match) RemovePersistentEffect(id int) {
	_, ok := match.persistentEffects[id]
	if ok {
		delete(match.persistentEffects, id)
	}
}
