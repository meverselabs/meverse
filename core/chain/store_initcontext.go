package chain

import (
	"io"
	"os"
)

func (st *Store) CopyContext(zipContextPath string) error {
	if in, err := os.Open(st.keydbPath); err != nil {
		return err
	} else {
		defer in.Close()

		st.closeLock.RLock()
		defer st.closeLock.RUnlock()
		if tempContextFile, err := os.Create(zipContextPath); err != nil {
			return err
		} else {
			if _, err = io.Copy(tempContextFile, in); err != nil {
				return err
			}
			if err = tempContextFile.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
