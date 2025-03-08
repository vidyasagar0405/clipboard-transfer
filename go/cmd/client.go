/*
Copyright © 2025 Vidyasagar Gopi vidyasagar0405@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"

	"secure-transfer/internal/crypto"
	"secure-transfer/internal/transfer"

	"github.com/spf13/cobra"
)

var (
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Send a file or message",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := crypto.GetAESKey(logger)
			if err != nil {
				return fmt.Errorf("error with encryption key: %w", err)
			}

			if cmd.Flags().Changed("message") {
				return transfer.SendMessage(ip, port, file, message, key, logger)
			}
			return transfer.SendFile(ip, port, file, key, logger)
		},
	}

	// Client-specific flags
	ip      string
	file    string
	message string
)

func init() {
	clientCmd.Flags().StringVarP(&ip, "ip", "i", "localhost", "Receiver IP address")
	clientCmd.Flags().StringVarP(&file, "file", "f", "", "File to send")
	clientCmd.Flags().StringVarP(&message, "message", "m", "", "Message to send instead of a file")
	clientCmd.MarkFlagRequired("ip")
}
