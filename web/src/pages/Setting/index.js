import React from 'react';
import { Layout } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import SystemSetting from '../../components/SystemSetting';
import OtherSetting from '../../components/OtherSetting';
import PersonalSetting from '../../components/PersonalSetting';
import OperationSetting from '../../components/OperationSetting';
import RateLimitSetting from '../../components/RateLimitSetting.js';
import ModelSetting from '../../components/ModelSetting.js';
import { isRoot } from '../../helpers';

const Setting = () => {
  const { t } = useTranslation();
  return (
    <div>
      <Layout>
        <Layout.Content>
          {isRoot() ? (
            <>
              <SystemSetting />
              <OperationSetting />
              <ModelSetting />
              <RateLimitSetting />
              <OtherSetting />
            </>
          ) : (
            <PersonalSetting />
          )}
        </Layout.Content>
      </Layout>
    </div>
  );
};

export default Setting;
