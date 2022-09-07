const apiPrefix = import.meta.env.VITE_PUBLIC_API_PREFIX ?? "/api";

export class APIError extends Error {
  constructor(
    public readonly message: string,
    public readonly statusCode: number
  ) {
    super(message);
  }
}

export function doAPIRequest<TR = unknown>(
  method: "POST" | "PUT" | "DELETE",
  path: string,
  expectedStatus: number,
  body?: Record<string, any>,
  noRedirect?: boolean
): Promise<TR>;
export function doAPIRequest<TR = unknown>(
  method: "GET",
  path: string,
  expectedStatus: number,
  body?: {},
  noRedirect?: boolean
): Promise<TR>;
export async function doAPIRequest<TR = unknown>(
  method: "GET" | "POST" | "PUT" | "DELETE",
  path: string,
  expectedStatus: number,
  body?: Record<string, any>,
  noRedirect = false
): Promise<TR> {
  let req;
  if (method === "GET") {
    req = fetch(`${apiPrefix}${path}`, {
      credentials: "include",
      headers: {
        Accept: "application/json",
      },
    });
  } else {
    req = fetch(`${apiPrefix}${path}`, {
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method,
      body: body && JSON.stringify(body),
    });
  }
  let res;
  try {
    res = await req;
  } catch (e) {
    console.error(`API FAIL ${method} ${path}`, e);
    throw e;
  }
  if (res.status !== expectedStatus) {
    console.error(
      `API ERROR ${method} ${path} ${res.status} != ${expectedStatus}`
    );
    const payload = await res.json();

    // Special-case for auth errors
    if (res.status === 401 && !noRedirect) {
      window.location.pathname = "/api/signIn";
      // @ts-expect-error - we're gone anyway
      return null;
    }

    let message: string;
    if ("error" in payload) {
      message = payload.error;
    } else if ("message" in payload) {
      message = payload.message;
    } else {
      message = JSON.stringify(payload);
    }
    console.error("\t" + message);
    throw new APIError(message, res.status);
  }
  return await res.json();
}
