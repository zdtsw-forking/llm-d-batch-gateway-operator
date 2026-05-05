package controller

import (
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/event"
)

// secretWatchFilter is a predicate that passes only Secret events for secrets
// the controller is actively tracking (i.e. cross-namespace secretRefs). It is
// updated dynamically during reconciliation so the cluster-wide Secret informer
// only triggers reconciles for secrets the controller actually cares about.
type secretWatchFilter struct {
	mu      sync.RWMutex
	watched map[string]struct{} // "namespace/name"
}

func (f *secretWatchFilter) add(ns, name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.watched[ns+"/"+name] = struct{}{}
}

func (f *secretWatchFilter) remove(ns, name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.watched, ns+"/"+name)
}

func (f *secretWatchFilter) matches(ns, name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.watched[ns+"/"+name]
	return ok
}

// Predicate interface — only Create/Update/Delete matter for secret rotation.
func (f *secretWatchFilter) Create(e event.CreateEvent) bool {
	return f.matches(e.Object.GetNamespace(), e.Object.GetName())
}
func (f *secretWatchFilter) Update(e event.UpdateEvent) bool {
	return f.matches(e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
}
func (f *secretWatchFilter) Delete(e event.DeleteEvent) bool {
	return f.matches(e.Object.GetNamespace(), e.Object.GetName())
}
func (f *secretWatchFilter) Generic(e event.GenericEvent) bool {
	return f.matches(e.Object.GetNamespace(), e.Object.GetName())
}
