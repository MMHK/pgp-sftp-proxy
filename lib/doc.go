//Package DahSing PGP files upload API
//
//	Schemes: http, https
//	Host: API_HOST
//	BasePath: /
//	Version: 1.0.1
//
//	Consumes:
//	 - multipart/form-data
//	 - application/json
//
//	Produces:
//	 - application/json
//
//	swagger:meta
package lib


type ResultResponse struct {
	// in: body
	Status bool   `json:"status"`
	Error  string `json:"error"`
}


// swagger:operation POST /encrypt encrypt
//
// Encrypt source file to PGP
//
// ---
// consumes:
//   - multipart/form-data
// produces:
//   - application/json
//   - text/plain; charset=utf-8
// parameters:
// - name: upload
//   type: file
//   in: formData
//   required: true
//   description: The file to upload.
// - name: key
//   type: string
//   in: formData
//   format: textarea
//   required: true
//   description: PGP public key
// responses:
//   200:
//     description: OK
//   500:
//     description: Error
//
//

// swagger:operation POST /upload upload
//
// Encrypt source file to PGP and Upload to SFTP
//
// ---
// consumes:
//   - multipart/form-data
// produces:
//   - application/json
// parameters:
// - name: upload
//   type: file
//   in: formData
//   required: true
//   description: The file to upload.
// - name: key
//   type: string
//   in: formData
//   required: true
//   format: textarea
//   description: PGP public key
// responses:
//   200:
//     description: OK
//   500:
//     description: Error


// swagger:operation GET /download upload
//
// Download PDF files of Policy from the SFTP
//
// ---
// consumes:
//   - application/json
// produces:
//   - application/json
// responses:
//   200:
//     description: OK
//   500:
//     description: Error

