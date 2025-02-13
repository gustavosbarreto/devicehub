import { AxiosError } from "axios";

function handleError(error: unknown) {
  if (process.env.NODE_ENV !== "test") {
    console.log(">>>>>>>>>>>>>>>>>>>>>");
    console.log(error);
    console.log(">>>>>>>>>>>>>>>>>>>>>");
    if (error instanceof AxiosError) throw new Error(`Axios error: ${error.status} - ${error.message}`);

    if (error instanceof Error) throw new Error(error.message);

    if (error instanceof Response) throw new Error(error.statusText);

    throw new Error("Unknown error");
  }
}

export default handleError;
