package recorder

/*
	Setup:	Delete all existing resources (env, triggers, functions, recorders)
			Create a python environment, hello function, trigger for hello function
			Establish a client connection to Redis pod in cluster

	Test1: 	Create a recorder for the hello function
			Issue a cURL request that should be recorded
			Check that it was by inspecting keys in Redis

	Test2:  Disable the recorder using the --disable flag
			Re-issue the cURL request
			Check that it was not recorded in Redis

	Test3:
 */
