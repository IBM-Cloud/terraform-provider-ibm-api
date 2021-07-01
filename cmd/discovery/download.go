// +build darwin

package main

import (
	"log"
	"os"
)

const tfrName = "terraformer"
const tfName = "terraform"

var goPath = os.Getenv("GOPATH")

// separate out command to install terraform and terraformer
// Add separate compile time files for each os - 3
// take versions (default to latest) and provide output as installed path
// ask user to add to the path

func downloadAndInitialize(tfrVersion, tfVersion, installPath string) error {
	log.Printf("versions: terraformer %s and terraform %s\n", tfrVersion, tfVersion)

	log.Println("TODO: Download terraformer and terraform here, if not available")

	if installPath != "" {
		if _, err := os.Stat(installPath); os.IsExist(err) {

		} else {

		}
	} else if goPath != "" {
		// Install in go path
		installPath = goPath
	} else {
		// Download into current folder and ask user to put it in path somewhere
		var err error
		installPath, err = os.Getwd()
		if err != nil {
			log.Println("Error with getting current folder, when downloading into current folder")
			return err
		}
	}

	// todo: @srikar - check existence of terraform
	if false {
		err := downloadTerraform(tfVersion)
		if err != nil {
			log.Println("Error with downloading terraform")
			return err
		}
	}

	// todo: @srikar - check existence of terraformer
	if false {
		// tfinst_bin="terraformer"
		// log.Println(" Terraformer as /go/bin/${tfinst_bin}")
		// #tfinst_url="https://api.github.com/repos/GoogleCloudPlatform/terraformer/releases/assets/${TERRAFORMER_ASSETID}"
		// tfinst_cmd="curl -LO https://github.com/GoogleCloudPlatform/terraformer/releases/download/${TERRAFORMER_VERSION}/terraformer-ibmcloud-linux-amd64"
		// tfinst_download_file "${tfinst_cmd}" ${TERRAFORMER_NAME}
		// chmod +x terraformer-ibmcloud-linux-amd64
		// mv terraformer-ibmcloud-linux-amd64 /go/bin/terraformer

		// log.Printf("Downloaded terraform successfully version: %s, installed at path %s",
		//  tfrVersion, installPath)

		// log.Printf("Downloaded terraformer successfully version: %s, installed at path %s",
		//  tfrVersion, installPath)
		err := downloadTar(tfrVersion, tfrGHRepo, "", "", true, true)
		if err != nil {
			log.Println("Error with downloading terraformer")
			return err
		}
	}

	//  // todo: @srikar -  where to place provider
	// todo: @srikar - check existence of terraform provider
	if false {
		err := downloadTar(tfProviderGHRepo, tfrGHRepo, "", "", true, true)
		if err != nil {
			log.Println("Error with downloading terraform provider")
			return err
		}
		// todo: @srikar - Place this in right place
	}

	return nil
}

func downloadTerraform(version string) error {
	return nil
}

// # installs terraform
// tfinst_bin="terraform"
// echo "\n### Installing Terraform as /go/bin/${tfinst_bin}"
// tfinst_zip="terraform_${TERRAFORM_VERSION}.zip"
// tfinst_url="${TERRAFORM_GIT_URL}/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
// tfinst_cmd="curl -v -L ${tfinst_url} -o ${tfinst_zip}"
// TERRAFORM_SHA256SUM=${TERRAFORM_11_SHA256SUM}

// if [ x"${TERRAFORM_VERSION}" = x"${TERRAFORM_12_VERSION}" ]; then
//   TERRAFORM_SHA256SUM=${TERRAFORM_12_SHA256SUM}

//   TERRAFORM13_SHA256SUM=${TERRAFORM_13_SHA256SUM}
//   tfinst13_bin="terraform13"
//   tfinst13_zip="terraform_${TERRAFORM_13_VERSION}.zip"
//   echo "\n### Installing Terraform 13 as /go/bin/${tfinst13_bin}"
//   tfinst13_url="${TERRAFORM_GIT_URL}/${TERRAFORM_13_VERSION}/terraform_${TERRAFORM_13_VERSION}_linux_amd64.zip"
//   tfinst13_cmd="curl -v -L ${tfinst13_url} -o ${tfinst13_zip}"
//   tfinst_download_file_and_verify_sha "${tfinst13_zip}" "${TERRAFORM_13_SHA256SUM}" "${tfinst13_cmd}"
//   tfinst_unzip_if_required ${tfinst13_zip}
//   mv terraform /go/bin/${tfinst13_bin}
//   if [ $? -ne 0 ]; then
//     echo "ERROR: While moving terraform 13 installer files"
//     exit 1
//   fi
//   # access given to terraform13 file
//   chmod +x /go/bin/${tfinst13_bin}

//   TERRAFORM14_SHA256SUM=${TERRAFORM_14_SHA256SUM}
//   tfinst14_bin="terraform14"
//   tfinst14_zip="terraform_${TERRAFORM_14_VERSION}.zip"
//   echo "\n### Installing Terraform 14 as /go/bin/${tfinst14_bin}"
//   tfinst14_url="${TERRAFORM_GIT_URL}/${TERRAFORM_14_VERSION}/terraform_${TERRAFORM_14_VERSION}_linux_amd64.zip"
//   tfinst14_cmd="curl -v -L ${tfinst14_url} -o ${tfinst14_zip}"
//   tfinst_download_file_and_verify_sha "${tfinst14_zip}" "${TERRAFORM_14_SHA256SUM}" "${tfinst14_cmd}"
//   tfinst_unzip_if_required ${tfinst14_zip}
//   mv terraform /go/bin/${tfinst14_bin}
//   if [ $? -ne 0 ]; then
//     echo "ERROR: While moving terraform 14 installer files"
//     exit 1
//   fi
//   # access given to terraform14 file
//   chmod +x /go/bin/${tfinst14_bin}

//   # dir for pre-installed plugins for terraform13
//   tfinst13_dir_preinstalled_tf_plugins=/home/nobody/.terraform.d/plugins/registry.terraform.io/ibm-cloud/ibm/${PROVIDER_IBM_12_VERSION}/linux_amd64
//   mkdir -p ${tfinst13_dir_preinstalled_tf_plugins}
//   echo "Terraform 13 and 14 installation done"

// fi

// # dir for pre-installed older plugins for terraform13
// tfinst13_dir_preinstalledibm_tf_plugins=/home/nobody/.terraform.d/plugins/registry.terraform.io/ibm-cloud/ibm

// tfinst_download_file_and_verify_sha "${tfinst_zip}" "${TERRAFORM_SHA256SUM}" "${tfinst_cmd}"
// tfinst_unzip_if_required ${tfinst_zip}
// mv terraform /go/bin/${tfinst_bin}
// if [ $? -ne 0 ] ; then
//   echo "ERROR: While moving terraform installer files"
//   exit 1
// fi
// #chown nobody:nogroup /go/bin/${tfinst_bin}
// chmod +x /go/bin/${tfinst_bin}

// # dir for pre-installed plugins
// tfinst_dir_preinstalled_tf_plugins=/home/nobody/.terraform.d/plugins
// mkdir -p ${tfinst_dir_preinstalled_tf_plugins}

// # installs the latest version of ibmcloud provider
// if [ x"${TERRAFORM_VERSION}" = x"${TERRAFORM_12_VERSION}" ] ; then
//   echo "\n### Installing IBM Cloud Provider as /home/nobody/.terraform.d/plugins/${PROVIDER_IBM_BIN}"
//   tfinst_url="${PROVIDER_IBM_GIT_URL}/${PROVIDER_IBM_12_GIT_ASSET_ID}"
//   tfinst_cmd="curl -v -L ${tfinst_url} -o ${PROVIDER_IBM_12_INSTALL_FILE} -H \"Accept: application/octet-stream\""
//   tfinst_download_file_and_verify_sha "${PROVIDER_IBM_12_INSTALL_FILE}" "${PROVIDER_IBM_12_SHA256SUM}" "${tfinst_cmd}"
//   tfinst_unzip_if_required ${PROVIDER_IBM_12_INSTALL_FILE}
//   cp ${PROVIDER_IBM_12_UNZIPPED_FILE} ${tfinst13_dir_preinstalled_tf_plugins}/${PROVIDER_IBM_12_BIN}
//   mv ${PROVIDER_IBM_12_UNZIPPED_FILE} ${tfinst_dir_preinstalled_tf_plugins}/${PROVIDER_IBM_12_BIN}
//   if [ $? -ne 0 ] ; then
//     echo "ERROR: While moving ibmcloud provider"
//     exit 1
//   fi
// else
//   echo "\n### Installing IBM Cloud Provider as /home/nobody/.terraform.d/plugins/${PROVIDER_IBM_BIN}"
//   tfinst_url="${PROVIDER_IBM_GIT_URL}/${PROVIDER_IBM_GIT_ASSET_ID}"
//   tfinst_cmd="curl -v -L ${tfinst_url} -o ${PROVIDER_IBM_INSTALL_FILE} -H \"Accept: application/octet-stream\""
//   tfinst_download_file_and_verify_sha "${PROVIDER_IBM_INSTALL_FILE}" "${PROVIDER_IBM_SHA256SUM}" "${tfinst_cmd}"
//   tfinst_unzip_if_required ${PROVIDER_IBM_INSTALL_FILE}
//   mv ${PROVIDER_IBM_UNZIPPED_FILE} ${tfinst_dir_preinstalled_tf_plugins}/${PROVIDER_IBM_BIN}
//   if [ $? -ne 0 ] ; then
//     echo "ERROR: While moving ibmcloud provider"
//     exit 1
//   fi
// fi
