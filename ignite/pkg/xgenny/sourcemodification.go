package xgenny

// SourceModification 描述運行後源代碼中修改和創建的文件
type SourceModification struct {
	modified map[string]struct{}
	created  map[string]struct{}
}

func NewSourceModification() SourceModification {
	return SourceModification{
		make(map[string]struct{}),
		make(map[string]struct{}),
	}
}

// ModifiedFiles返回源修改的修改文件
func (sm SourceModification) ModifiedFiles() (modifiedFiles []string) {
	for modified := range sm.modified {
		modifiedFiles = append(modifiedFiles, modified)
	}
	return
}

// CreatedFiles返回源修改創建的文件
func (sm SourceModification) CreatedFiles() (createdFiles []string) {
	for created := range sm.created {
		createdFiles = append(createdFiles, created)
	}
	return
}

// AppendModifiedFiles在源修改中附加尚未記錄的修改文件
func (sm *SourceModification) AppendModifiedFiles(modifiedFiles ...string) {
	for _, modifiedFile := range modifiedFiles {
		_, alreadyModified := sm.modified[modifiedFile]
		_, alreadyCreated := sm.created[modifiedFile]
		if !alreadyModified && !alreadyCreated {
			sm.modified[modifiedFile] = struct{}{}
		}
	}
}

// AppendCreatedFiles在尚未記錄的源修改中附加創建的文件
func (sm *SourceModification) AppendCreatedFiles(createdFiles ...string) {
	for _, createdFile := range createdFiles {
		_, alreadyModified := sm.modified[createdFile]
		_, alreadyCreated := sm.created[createdFile]
		if !alreadyModified && !alreadyCreated {
			sm.created[createdFile] = struct{}{}
		}
	}
}

// 合併將新的源修改合併到現有的修改
func (sm *SourceModification) Merge(newSm SourceModification) {
	sm.AppendModifiedFiles(newSm.ModifiedFiles()...)
	sm.AppendCreatedFiles(newSm.CreatedFiles()...)
}
