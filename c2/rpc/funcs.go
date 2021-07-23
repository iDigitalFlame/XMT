package rpc

/*

	GET /api/v1/session
		JSON list of Sessions
	GET /api/v1/session/<session_id>
		JSON info for Session with the provided ID
	GET /session/<session_id>/job
		JSON list of all Jobs for the Session
	PUT /session/<session_id>/exec
		Body
			{
				"cmd": <str, required, greater than size 2>,
				"wait": <boolean, optional>
			}
		Result is a JSON info for the created Job
	GET /session/<session_id>/<job_id>
		JSON info for the Session Job with specified ID
	DELETE /session/<session_id>/<job_id>
		Remove the specified Job from cache
	GET /session/<session_id>/<job_id>/result
		Binary output of the Job data, 404 unless completed

	GET /listener
		JSON list of Listeners
	GET /listener/<listener_id>
		JSON info for Listener along with connected Sessions
*/
