/* eslint-disable @typescript-eslint/no-explicit-any */
import dbmodels from '@/lib/dbModels';
import type { ServiceConfig as OriginalServiceConfig } from '@/lib/types';

type ServiceConfig = Omit<OriginalServiceConfig, '_id'>;

const getLastZoneFetchFromDB = async (): Promise<ServiceConfig | null> => {
    const { serviceConfigModel } = dbmodels;
    const config = await serviceConfigModel.findOne({});

    if (!config) {
        return null;
    }

    const configObj = config.toObject();
    delete (configObj as any)._id;
    return configObj;
};

export const checkLastZoneFetch = async (): Promise<Date | string | null> => {
    const lastFetch = await getLastZoneFetchFromDB();
    
    if (!lastFetch) {
        return null;
    }

    console.log({ lastFetch });
    return lastFetch.lastZoneFetch || null;
};
