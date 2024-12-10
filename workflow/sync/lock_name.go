package sync

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj/argo-workflows/v3/errors"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
)

type lockKind string

const (
	lockKindConfigMap lockKind = "ConfigMap"
	lockKindMutex     lockKind = "Mutex"
)

type lockName struct {
	Namespace    string
	ResourceName string
	Key          string
	Kind         lockKind
	Selectors    []v1alpha1.SyncSelector
}

func newLockName(namespace, resourceName, lockKey string, kind lockKind, selectors []v1alpha1.SyncSelector) *lockName {
	return &lockName{
		Namespace:    namespace,
		ResourceName: resourceName,
		Key:          lockKey,
		Kind:         kind,
		Selectors:    selectors,
	}
}

func getSemaphoreLockName(sem *v1alpha1.SemaphoreRef, wfNamespace string) (*lockName, error) {
	if sem.ConfigMapKeyRef != nil {
		namespace := sem.Namespace
		if namespace == "" {
			namespace = wfNamespace
		}
		return newLockName(namespace, sem.ConfigMapKeyRef.Name, sem.ConfigMapKeyRef.Key, lockKindConfigMap, sem.Selectors), nil
	}
	return nil, fmt.Errorf("cannot get LockName for a Semaphore without a ConfigMapRef")
}

func getMutexLockName(mtx *v1alpha1.Mutex, wfNamespace string) *lockName {
	namespace := mtx.Namespace
	if namespace == "" {
		namespace = wfNamespace
	}
	return newLockName(namespace, mtx.Name, "", lockKindMutex, mtx.Selectors)
}

func getLockName(item *syncItem, wfNamespace string) (*lockName, error) {
	switch {
	case item.semaphore != nil:
		return getSemaphoreLockName(item.semaphore, wfNamespace)
	case item.mutex != nil:
		return getMutexLockName(item.mutex, wfNamespace), nil
	default:
		return nil, fmt.Errorf("cannot getLockName if not semaphore or mutex")
	}
}

func DecodeLockName(name string) (*lockName, error) {
	splittedLockName := strings.Split(name, "?")
	lockNameTrimedSelectors := splittedLockName[0]
	selectors := ParseSelectors(strings.Join(splittedLockName[1:], "?"))
	items := strings.SplitN(lockNameTrimedSelectors, "/", 3)
	if len(items) < 3 {
		return nil, errors.New(errors.CodeBadRequest, "Invalid lock key: unknown format")
	}

	var lock lockName
	lockKind := lockKind(items[1])
	namespace := items[0]

	switch lockKind {
	case lockKindMutex:
		lock = lockName{Namespace: namespace, Kind: lockKind, ResourceName: items[2], Selectors: selectors}
	case lockKindConfigMap:
		components := strings.Split(items[2], "/")

		if len(components) != 2 {
			return nil, errors.New(errors.CodeBadRequest, "Invalid ConfigMap lock key: unknown format")
		}

		lock = lockName{Namespace: namespace, Kind: lockKind, ResourceName: components[0], Key: components[1], Selectors: selectors}
	default:
		return nil, errors.New(errors.CodeBadRequest, fmt.Sprintf("Invalid lock key, unexpected kind: %s", lockKind))
	}

	err := lock.validate()
	if err != nil {
		return nil, err
	}
	return &lock, nil
}

func StringifySelectors(selectors []v1alpha1.SyncSelector) string {
	joinedSelectors := ""
	for _, selector := range selectors {
		// at this point template should be already replaced
		if selector.Template != "" {
			// escape & and = chars to decode easily later
			re := regexp.MustCompile("&|=")
			escapedSelectorName := re.ReplaceAllString(selector.Name, "-")
			escapedSelectorValue := re.ReplaceAllString(selector.Template, "-")

			joinedSelectors = joinedSelectors + fmt.Sprintf("%s=%s&", escapedSelectorName, escapedSelectorValue)
		}
	}
	return strings.TrimRight(joinedSelectors, "&")
}

func ParseSelectors(selectors string) []v1alpha1.SyncSelector {
	parsedSelectors := []v1alpha1.SyncSelector{}
	splittedSelectors := strings.Split(selectors, "&")

	for _, selectorStr := range splittedSelectors {
		keyValPair := strings.Split(selectorStr, "=")
		if len(keyValPair) == 2 {
			parsedSelectors = append(parsedSelectors, v1alpha1.SyncSelector{
				Name:     keyValPair[0],
				Template: keyValPair[1],
			})
		}
		// otherwise consider invalid, do nothing
	}
	return parsedSelectors
}

func (ln *lockName) encodeName() string {
	encodingBuilder := &strings.Builder{}

	encodingBuilder.WriteString(fmt.Sprintf("%s/%s/%s", ln.Namespace, ln.Kind, ln.ResourceName))
	if ln.Kind == lockKindConfigMap {
		encodingBuilder.WriteString(fmt.Sprintf("/%s", ln.Key))
	}
	if selectors := StringifySelectors(ln.Selectors); len(selectors) > 0 {
		encodingBuilder.WriteString(fmt.Sprintf("?%s", selectors))
	}
	return ln.validateEncoding(encodingBuilder.String())
}

func (ln *lockName) validate() error {
	if ln.Namespace == "" {
		return errors.New(errors.CodeBadRequest, "Invalid lock key: Namespace is missing")
	}
	if ln.Kind == "" {
		return errors.New(errors.CodeBadRequest, "Invalid lock key: Kind is missing")
	}
	if ln.ResourceName == "" {
		return errors.New(errors.CodeBadRequest, "Invalid lock key: ResourceName is missing")
	}
	if ln.Kind == lockKindConfigMap && ln.Key == "" {
		return errors.New(errors.CodeBadRequest, "Invalid lock key: Key is missing for ConfigMap lock")
	}
	return nil
}

func (ln *lockName) validateEncoding(encoding string) string {
	decoded, err := DecodeLockName(encoding)
	if err != nil {
		panic(fmt.Sprintf("bug: unable to decode lock that was just encoded: %s", err))
	}
	if ln.Namespace != decoded.Namespace || ln.Kind != decoded.Kind || ln.ResourceName != decoded.ResourceName || ln.Key != decoded.Key {
		panic("bug: lock that was just encoded does not match encoding")
	}
	return encoding
}
