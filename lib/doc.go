//Package Zurich PGP files upload API
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

// swagger:response ResultResponse
type ResultResponse struct {
	// in: body
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

// swagger:parameters hello
type JSONBody struct {
	Body *MultipleBody
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
// - name: deploy
//   type: string
//   in: formData
//   required: true
//   enum: [dev, pro, test]
//   description: sftp remote save folder
// responses:
//   200:
//     description: OK
//   500:
//     description: Error

// swagger:operation POST /multiple/upload multipleUpload
//
// Encrypt source file to PGP and Upload to SFTP
//
// ---
// consumes:
//   - application/json
// produces:
//   - application/json
// parameters:
// - in: body
//   name: body
//   description: request body
//   schema:
//	   "$ref": "#/definitions/MultipleBody"
// responses:
//   200:
//     description: OK
//   500:
//     description: Error
