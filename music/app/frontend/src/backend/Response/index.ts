export enum ResponseStatus {
  Success = 'success',
  Error = 'error',
}

export interface Response {
  status: ResponseStatus;
  message: any;
}
