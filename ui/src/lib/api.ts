const apiPrefix = import.meta.env.VITE_PUBLIC_API_PREFIX ?? "/api";

export class APIError extends Error {
  constructor(
    public readonly message: string,
    public readonly statusCode: number
  ) {
    super(message);
  }
}

export function doAPIRequest(
  method: "POST" | "PUT" | "DELETE",
  path: string,
  expectedStatus: number,
  body?: Record<string, any>,
  noRedirect?: boolean
): Promise<unknown>;
export function doAPIRequest(
  method: "GET",
  path: string,
  expectedStatus: number,
  body?: {},
  noRedirect?: boolean
): Promise<unknown>;
export async function doAPIRequest(
  method: "GET" | "POST" | "PUT" | "DELETE",
  path: string,
  expectedStatus: number,
  body?: Record<string, any>,
  noRedirect = false
): Promise<unknown> {
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
      return null;
    }

    let message: string;
    if ("error" in payload) {
      message = payload.error;
    } else {
      message = JSON.stringify(payload);
    }
    console.error("\t" + message);
    throw new APIError(message, res.status);
  }
  return await res.json();
}
