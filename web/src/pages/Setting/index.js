import React, { useState } from 'react';
import { Layout, Nav, Typography, Card } from '@douyinfe/semi-ui';
import { 
  IconSetting, 
  IconUser, 
  IconCustomize,
  IconPieChartStroked,
  IconServer
} from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import SystemSetting from '../../components/SystemSetting';
import OtherSetting from '../../components/OtherSetting';
import PersonalSetting from '../../components/PersonalSetting';
import OperationSetting from '../../components/OperationSetting';
import RateLimitSetting from '../../components/RateLimitSetting.js';
import ModelSetting from '../../components/ModelSetting.js';
import { isRoot } from '../../helpers';
import './style.css';

const { Header, Sider, Content } = Layout;
const { Title } = Typography;

const Setting = () => {
  const { t } = useTranslation();
  const [activeKey, setActiveKey] = useState(isRoot() ? 'system' : 'personal');
  
  const handleNavChange = (key) => {
    setActiveKey(key);
  };
  
  const renderSettingComponent = () => {
    switch (activeKey) {
      case 'system':
        return <SystemSetting />;
      case 'operation':
        return <OperationSetting />;
      case 'model':
        return <ModelSetting />;
      case 'ratelimit':
        return <RateLimitSetting />;
      case 'other':
        return <OtherSetting />;
      case 'personal':
      default:
        return <PersonalSetting />;
    }
  };
  
  return (
    <div className="setting-container">
      <Layout className="setting-layout">
        <Sider className="setting-sider">
          <div className="setting-logo">
            <IconSetting size="large" />
            <span>{t('设置')}</span>
          </div>
          <Nav
            defaultSelectedKeys={[activeKey]}
            onSelect={({ itemKey }) => handleNavChange(itemKey)}
            items={isRoot() ? [
              { itemKey: 'system', text: t('系统设置'), icon: <IconServer /> },
              { itemKey: 'operation', text: t('运营设置'), icon: <IconSetting /> },
              { itemKey: 'model', text: t('模型设置'), icon: <IconCustomize /> },
              { itemKey: 'ratelimit', text: t('限流设置'), icon: <IconPieChartStroked /> },
              { itemKey: 'other', text: t('其他设置'), icon: <IconSetting /> },
              { itemKey: 'personal', text: t('个人设置'), icon: <IconUser /> }
            ] : [
              { itemKey: 'personal', text: t('个人设置'), icon: <IconUser /> }
            ]}
            style={{ height: '100%' }}
          />
        </Sider>
        <Layout>
          <Header className="setting-header">
            <Title heading={3} style={{ margin: 0 }}>
              {activeKey === 'system' && t('系统设置')}
              {activeKey === 'operation' && t('运营设置')}
              {activeKey === 'model' && t('模型设置')}
              {activeKey === 'ratelimit' && t('限流设置')}
              {activeKey === 'other' && t('其他设置')}
              {activeKey === 'personal' && t('个人设置')}
            </Title>
          </Header>
          <Content className="setting-content">
            <div className="setting-scroll-container">
              {renderSettingComponent()}
            </div>
          </Content>
        </Layout>
      </Layout>
    </div>
  );
};

export default Setting;
