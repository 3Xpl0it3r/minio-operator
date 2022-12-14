/*
   Copyright 2022 The minio-operator Authors.
   Licensed under the Apache License, PROJECT_VERSION 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package options

import (
	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"
)

type Options struct {
	// this is example flags
	ListenAddress string
	// todo write your flags here
}

var _ options = new(Options)

// NewOptions create an instance option and return
func NewOptions() *Options {
	// todo write your code or change this code here
	return &Options{}
}

// Validate validates options
func (o *Options) Validate() []error {
	// todo write your code here, if you need some validation
	return nil
}

// Complete fill some default value to options
func (o *Options) Complete() error {
	// todo write your code here, you may do some defaulter if neceressary
	return nil
}

//
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ListenAddress, "web.listen-addr", ":8080", "Address on which to expose metrics and web interfaces")
	// todo write your code here
}

func (o *Options) NamedFlagSets() (fs flag.NamedFlagSets) {
	o.AddFlags(fs.FlagSet("minio-operator"))
	// other options addFlags
	return
}
