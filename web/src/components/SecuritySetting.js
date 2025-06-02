import React, { useEffect, useState } from 'react';
import { Layout, Nav, Card, Typography, Spin } from '@douyinfe/semi-ui';
import { 
  IconShield, 
  IconLock, 
  IconMonitor, 
  IconAlertTriangle,
  IconSettings,
  IconList
} from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../helpers';
import SecurityOverview from './Security/SecurityOverview';
import AbnormalDetectionSettings from './Security/AbnormalDetectionSettings';
import RequestLimitSettings from './Security/RequestLimitSettings';
import StreamProtectionSettings from './Security/StreamProtectionSettings';
import IPBlacklistSettings from './Security/IPBlacklistSettings';
import SecurityLogs from './Security/SecurityLogs';

const { Sider, Content } = Layout;
const { Title } = Typography;

const SecuritySetting = () => {
  const { t } = useTranslation();
  const [activeKey, setActiveKey] = useState('overview');
  const [loading, setLoading] = useState(false);
  const [securityConfig, setSecurityConfig] = useState(null);
  const [securityStats, setSecurityStats] = useState(null);

  const loadSecurityConfig = async () => {
    setLoading(true);
    try {
      const [configRes, statsRes] = await Promise.all([
        API.get('/api/security/config'),
        API.get('/api/security/stats')
      ]);
      
      if (configRes.data.success) {
        setSecurityConfig(configRes.data.data);
      }
      
      if (statsRes.data.success) {
        setSecurityStats(statsRes.data.data);
      }
    } catch (error) {
      showError(t('加载安全配置失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSecurityConfig();
  }, []);

  const handleNavChange = (key) => {
    setActiveKey(key);
  };

  const refreshData = () => {
    loadSecurityConfig();
  };

  const renderSettingComponent = () => {
    const commonProps = {
      config: securityConfig,
      stats: securityStats,
      refresh: refreshData,
      loading
    };

    switch (activeKey) {
      case 'overview':
        return <SecurityOverview {...commonProps} />;
      case 'abnormal':
        return <AbnormalDetectionSettings {...commonProps} />;
      case 'request':
        return <RequestLimitSettings {...commonProps} />;
      case 'stream':
        return <StreamProtectionSettings {...commonProps} />;
      case 'blacklist':
        return <IPBlacklistSettings {...commonProps} />;
      case 'logs':
        return <SecurityLogs {...commonProps} />;
      default:
        return <SecurityOverview {...commonProps} />;
    }
  };

  const getPageTitle = () => {
    const titles = {
      overview: t('安全概览'),
      abnormal: t('异常检测'),
      request: t('请求限制'),
      stream: t('流保护'),
      blacklist: t('IP管理'),
      logs: t('安全日志')
    };
    return titles[activeKey] || t('安全设置');
  };

  return (
    <div style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Layout style={{ flex: 1, background: 'var(--semi-color-bg-0)' }}>
        <Sider 
          style={{ 
            background: 'var(--semi-color-bg-1)',
            borderRight: '1px solid var(--semi-color-border)'
          }}
          width={240}
        >
          <div style={{ 
            padding: '20px 16px', 
            borderBottom: '1px solid var(--semi-color-border)',
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}>
            <IconShield size="large" style={{ color: 'var(--semi-color-primary)' }} />
            <Title heading={4} style={{ margin: 0 }}>{t('安全设置')}</Title>
          </div>
          
          <Nav
            selectedKeys={[activeKey]}
            onSelect={({ itemKey }) => handleNavChange(itemKey)}
            items={[
              { 
                itemKey: 'overview', 
                text: t('安全概览'), 
                icon: <IconMonitor /> 
              },
              { 
                itemKey: 'abnormal', 
                text: t('异常检测'), 
                icon: <IconAlertTriangle /> 
              },
              { 
                itemKey: 'request', 
                text: t('请求限制'), 
                icon: <IconSettings /> 
              },
              { 
                itemKey: 'stream', 
                text: t('流保护'), 
                icon: <IconShield /> 
              },
              { 
                itemKey: 'blacklist', 
                text: t('IP管理'), 
                icon: <IconLock /> 
              },
              { 
                itemKey: 'logs', 
                text: t('安全日志'), 
                icon: <IconList /> 
              }
            ]}
            style={{ 
              marginTop: '16px',
              padding: '0 8px'
            }}
          />
        </Sider>
        
        <Layout>
          <Content style={{ 
            padding: '24px',
            background: 'var(--semi-color-bg-0)',
            overflow: 'auto'
          }}>
            <div style={{ marginBottom: '24px' }}>
              <Title heading={3} style={{ margin: 0 }}>
                {getPageTitle()}
              </Title>
            </div>
            
            <Spin spinning={loading}>
              {renderSettingComponent()}
            </Spin>
          </Content>
        </Layout>
      </Layout>
    </div>
  );
};

export default SecuritySetting;
