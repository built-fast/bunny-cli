package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newStorageCmd() *cobra.Command {
	var (
		password string
		hostname string
	)

	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage files in Edge Storage",
	}

	cmd.PersistentFlags().StringVar(&password, "password", "", "Storage zone password (auto-detected from zone if omitted)")
	cmd.PersistentFlags().StringVar(&hostname, "hostname", "", "Storage API hostname (auto-detected from zone if omitted)")

	cmd.AddCommand(newStorageLsCmd())
	cmd.AddCommand(newStorageCpCmd())
	cmd.AddCommand(newStorageRmCmd())

	return cmd
}

// parseZonePath splits "zone/path/to/file" into zone name and path.
// If there is no "/", path is empty (root of zone).
func parseZonePath(arg string) (zoneName, path string) {
	idx := strings.IndexByte(arg, '/')
	if idx < 0 {
		return arg, ""
	}
	return arg[:idx], arg[idx+1:]
}

// resolveStorageCredentials returns the password and hostname for a storage zone.
// If explicit flags are provided, they are used. Otherwise, the zone is looked up
// via the main API by name.
func resolveStorageCredentials(cmd *cobra.Command, zoneName string) (password, hostname string, err error) {
	password, _ = cmd.Flags().GetString("password")
	hostname, _ = cmd.Flags().GetString("hostname")

	if password != "" && hostname != "" {
		return password, hostname, nil
	}

	// Auto-lookup via main API
	szAPI, err := AppFromContext(cmd.Context()).NewStorageZoneAPI(cmd)
	if err != nil {
		return "", "", fmt.Errorf("looking up storage zone: %w", err)
	}

	sz, err := szAPI.FindStorageZoneByName(cmd.Context(), zoneName)
	if err != nil {
		return "", "", err
	}

	if password == "" {
		password = sz.Password
	}
	if hostname == "" {
		hostname = sz.StorageHostname
	}

	return password, hostname, nil
}

func newStorageLsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls <zone>[/<path>]",
		Short: "List files and directories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			zoneName, path := parseZonePath(args[0])

			password, hostname, err := resolveStorageCredentials(cmd, zoneName)
			if err != nil {
				return err
			}

			storageAPI, err := AppFromContext(cmd.Context()).NewStorageAPI(cmd, password, hostname)
			if err != nil {
				return err
			}

			cfg := output.FromContext(cmd.Context())

			objects, err := storageAPI.ListFiles(cmd.Context(), zoneName, path)
			if err != nil {
				return err
			}

			columns := storageObjectColumns()

			if cfg.HasFields() {
				if err := output.ValidateFields(columns, cfg.Fields); err != nil {
					return err
				}
			}

			items := make([]any, len(objects))
			for i := range objects {
				items[i] = &objects[i]
			}

			formatted, err := output.FormatList(cfg, columns, items, false)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		},
	}

	return cmd
}

func newStorageCpCmd() *cobra.Command {
	var checksum bool

	cmd := &cobra.Command{
		Use:   "cp <src> <dst>",
		Short: "Upload or download files",
		Long:  "Copy files to/from Edge Storage. If src is a local file, uploads it. Otherwise, downloads from storage.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst := args[0], args[1]

			// Direction detection: if src exists as a local file, upload
			_, statErr := os.Stat(src)
			if statErr == nil {
				return runUpload(cmd, src, dst, checksum)
			}

			// Otherwise, download
			return runDownload(cmd, src, dst)
		},
	}

	cmd.Flags().BoolVar(&checksum, "checksum", false, "Compute and send SHA256 checksum on upload")

	return cmd
}

func runUpload(cmd *cobra.Command, localPath, remotePath string, sendChecksum bool) error {
	zoneName, path := parseZonePath(remotePath)
	if path == "" {
		return fmt.Errorf("remote path must include a file path (e.g., %s/path/file.txt)", zoneName)
	}

	password, hostname, err := resolveStorageCredentials(cmd, zoneName)
	if err != nil {
		return err
	}

	storageAPI, err := AppFromContext(cmd.Context()).NewStorageAPI(cmd, password, hostname)
	if err != nil {
		return err
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("reading file info: %w", err)
	}
	size := info.Size()

	var checksumHex string
	if sendChecksum {
		// Compute SHA256, then seek back to start
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return fmt.Errorf("computing checksum: %w", err)
		}
		checksumHex = fmt.Sprintf("%X", h.Sum(nil))
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("seeking file: %w", err)
		}
	}

	var reader io.Reader = f
	if showProgress(cmd) {
		bar := progressbar.DefaultBytes(size, "uploading")
		reader = io.TeeReader(f, bar)
		defer func() { _ = bar.Close() }()
	}

	if err := storageAPI.UploadFile(cmd.Context(), zoneName, path, reader, size, checksumHex); err != nil {
		return err
	}

	if !showProgress(cmd) {
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Upload complete.")
	} else {
		fmt.Fprintln(cmd.ErrOrStderr())
	}
	return err
}

func runDownload(cmd *cobra.Command, remotePath, localPath string) error {
	zoneName, path := parseZonePath(remotePath)
	if path == "" {
		return fmt.Errorf("remote path must include a file path (e.g., %s/path/file.txt)", zoneName)
	}

	password, hostname, err := resolveStorageCredentials(cmd, zoneName)
	if err != nil {
		return err
	}

	storageAPI, err := AppFromContext(cmd.Context()).NewStorageAPI(cmd, password, hostname)
	if err != nil {
		return err
	}

	body, contentLength, err := storageAPI.DownloadFile(cmd.Context(), zoneName, path)
	if err != nil {
		return err
	}
	defer func() { _ = body.Close() }()

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	var writer io.Writer = outFile
	if showProgress(cmd) && contentLength > 0 {
		bar := progressbar.DefaultBytes(contentLength, "downloading")
		writer = io.MultiWriter(outFile, bar)
		defer func() {
			_ = bar.Close()
			fmt.Fprintln(cmd.ErrOrStderr())
		}()
	}

	if _, err := io.Copy(writer, body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	if !showProgress(cmd) || contentLength <= 0 {
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Download complete.")
	}
	return err
}

func newStorageRmCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "rm <zone>/<path>",
		Short: "Delete a file or directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			zoneName, path := parseZonePath(args[0])
			if path == "" {
				return fmt.Errorf("path is required (e.g., %s/path/file.txt)", zoneName)
			}

			if !yes {
				confirmed, err := confirm(cmd, fmt.Sprintf("Are you sure you want to delete %s/%s? [y/N] ", zoneName, path))
				if err != nil {
					return err
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Deletion canceled.")
					return err
				}
			}

			password, hostname, err := resolveStorageCredentials(cmd, zoneName)
			if err != nil {
				return err
			}

			storageAPI, err := AppFromContext(cmd.Context()).NewStorageAPI(cmd, password, hostname)
			if err != nil {
				return err
			}

			if err := storageAPI.DeleteFile(cmd.Context(), zoneName, path); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "File deleted.")
			return err
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

// showProgress returns true if a progress bar should be shown.
func showProgress(cmd *cobra.Command) bool {
	cfg := output.FromContext(cmd.Context())
	if isJSONFormat(cfg.Format) {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// storageObjectColumns defines the columns for storage object list output.
func storageObjectColumns() []output.Column {
	return output.ToColumns([]output.TypedColumn[*client.StorageObject]{
		output.StringColumn[*client.StorageObject]("Name", func(o *client.StorageObject) string { return o.ObjectName }),
		output.StringColumn[*client.StorageObject]("Type", func(o *client.StorageObject) string {
			if o.IsDirectory {
				return "dir"
			}
			return "file"
		}),
		output.IntColumn[*client.StorageObject]("Size", func(o *client.StorageObject) int { return int(o.Length) }),
		output.StringColumn[*client.StorageObject]("Last Changed", func(o *client.StorageObject) string { return o.LastChanged }),
		output.StringColumn[*client.StorageObject]("Date Created", func(o *client.StorageObject) string { return o.DateCreated }),
	})
}
