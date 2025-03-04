console.log('main function started');

Deno.serve(async (req: Request) => {
  const url = new URL(req.url);
  const {pathname} = url;
  console.log(`serving request for ${pathname}`);

  // handle health checks
  if (pathname === '/_internal/health') {
    return new Response(
      JSON.stringify({'message': 'ok'}),
      {status: 200, headers: {'Content-Type': 'application/json'}},
    );
  }

  if (pathname === '/_internal/metric') {
    const metric = await EdgeRuntime.getRuntimeMetrics();
    return Response.json(metric);
  }

  if (!pathname || pathname === '') {
    const error = {msg: 'missing function name in request'};
    return new Response(
      JSON.stringify(error),
      {status: 400, headers: {'Content-Type': 'application/json'}},
    );
  }


  const servicePath = `.${pathname}`;
  console.log(`serving the request with ${servicePath}`);

  const createWorker = async () => {
    const memoryLimitMb = 150; // TODO: load from shaple builder config
    const workerTimeoutMs = 5 * 60 * 1000; // TODO: load from shaple builder config
    const noModuleCache = false; // TODO: load from shaple builder config

    // TODO: set import map when file exists
    // const importMapPath = `data:${encodeURIComponent(JSON.stringify(importMap))}?${encodeURIComponent('/home/deno/functions/test')}`;
    const importMapPath = null;
    const envVarsObj = Deno.env.toObject();
    const envVars = Object.keys(envVarsObj).map((k) => [k, envVarsObj[k]]);
    const forceCreate = false;
    const netAccessDisabled = false;

    const cpuTimeSoftLimitMs = 10000;
    const cpuTimeHardLimitMs = 20000;

    return await EdgeRuntime.userWorkers.create({
      servicePath,
      memoryLimitMb,
      workerTimeoutMs,
      noModuleCache,
      importMapPath,
      envVars,
      forceCreate,
      netAccessDisabled,
      cpuTimeSoftLimitMs,
      cpuTimeHardLimitMs,
    });
  };

  const callWorker = async () => {
    try {
      // If a worker for the given service path already exists,
      // it will be reused by default.
      // Update forceCreate option in createWorker to force create a new worker for each request.
      const worker = await createWorker();
      const controller = new AbortController();

      const signal = controller.signal;
      // Optional: abort the request after a timeout
      //setTimeout(() => controller.abort(), 2 * 60 * 1000);

      return await worker.fetch(req, {signal});
    } catch (e) {
      console.error(e);

      if (e instanceof Deno.errors.WorkerRequestCancelled) {
        // XXX(Nyannyacha): I can't think right now how to re-poll
        // inside the worker pool without exposing the error to the
        // surface.

        // It is satisfied when the supervisor that handled the original
        // request terminated due to reaches such as CPU time limit or
        // Wall-clock limit.
        //
        // The current request to the worker has been canceled due to
        // some internal reasons. We should repoll the worker and call
        // `fetch` again.
        // return await callWorker();
        console.log('cancelled!');
      }

      const error = {msg: e.toString()};
      return new Response(
        JSON.stringify(error),
        {status: 500, headers: {'Content-Type': 'application/json'}},
      );
    }
  };

  return callWorker();
});