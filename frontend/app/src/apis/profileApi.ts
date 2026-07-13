import { httpClientWithAuth } from './httpClient';

export const profileApi = {
  registerPushToken: async (token: string): Promise<void> => {
    await httpClientWithAuth.putWithToken('/app/profile/push-token', { token });
  },
};
