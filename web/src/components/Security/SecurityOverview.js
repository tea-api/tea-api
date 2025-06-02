import React from 'react';
import { Card, Row, Col, Typography, Tag, Progress, Space, Descriptions, Button } from '@douyinfe/semi-ui';
import { 
  IconShield, 
  IconAlertTriangle, 
  IconLock, 
  IconMonitor,
  IconRefresh
} from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

const SecurityOverview = ({ stats, config, refresh }) => {
  const { t } = useTranslation();

  if (!stats || !config) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '40px' }}>
          <Text>{t('加载中...')}</Text>
        </div>
      </Card>
    );
  }

  const securityStats = stats.security_stats || {};
  const blacklistStats = stats.blacklist_stats || {};
  const streamStats = stats.stream_stats || {};

  // 计算安全评分
  const calculateSecurityScore = () => {
    let score = 100;
    
    // 根据被阻止的请求数量扣分
    if (securityStats.blocked_requests > 100) score -= 20;
    else if (securityStats.blocked_requests > 50) score -= 10;
    
    // 根据恶意检测数量扣分
    if (securityStats.malicious_detections > 50) score -= 15;
    else if (securityStats.malicious_detections > 20) score -= 8;
    
    // 根据黑名单IP数量扣分
    if (blacklistStats.total_blacklisted > 20) score -= 10;
    else if (blacklistStats.total_blacklisted > 10) score -= 5;
    
    return Math.max(score, 0);
  };

  const securityScore = calculateSecurityScore();
  
  const getScoreColor = (score) => {
    if (score >= 90) return 'green';
    if (score >= 70) return 'orange';
    return 'red';
  };

  const getScoreStatus = (score) => {
    if (score >= 90) return t('安全');
    if (score >= 70) return t('警告');
    return t('危险');
  };

  return (
    <div>
      {/* 安全评分卡片 */}
      <Card style={{ marginBottom: '24px' }}>
        <Row gutter={24}>
          <Col span={8}>
            <div style={{ textAlign: 'center' }}>
              <IconShield 
                size="extra-large" 
                style={{ 
                  color: getScoreColor(securityScore) === 'green' ? '#52c41a' : 
                         getScoreColor(securityScore) === 'orange' ? '#fa8c16' : '#ff4d4f',
                  marginBottom: '16px'
                }} 
              />
              <Title heading={2} style={{ margin: '8px 0' }}>
                {securityScore}
              </Title>
              <Tag 
                color={getScoreColor(securityScore)}
                size="large"
              >
                {getScoreStatus(securityScore)}
              </Tag>
            </div>
          </Col>
          <Col span={16}>
            <Title heading={4} style={{ marginBottom: '16px' }}>
              {t('安全状态概览')}
            </Title>
            <Progress 
              percent={securityScore} 
              stroke={getScoreColor(securityScore)}
              showInfo={false}
              style={{ marginBottom: '16px' }}
            />
            <Space>
              <Button 
                icon={<IconRefresh />} 
                onClick={refresh}
                type="tertiary"
              >
                {t('刷新数据')}
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 统计数据卡片 */}
      <Row gutter={24} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <IconAlertTriangle 
                size="large" 
                style={{ color: '#fa8c16', marginBottom: '8px' }} 
              />
              <Title heading={3} style={{ margin: '8px 0' }}>
                {securityStats.blocked_requests || 0}
              </Title>
              <Text type="secondary">{t('被阻止的请求')}</Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <IconLock 
                size="large" 
                style={{ color: '#ff4d4f', marginBottom: '8px' }} 
              />
              <Title heading={3} style={{ margin: '8px 0' }}>
                {securityStats.malicious_detections || 0}
              </Title>
              <Text type="secondary">{t('恶意行为检测')}</Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <IconShield 
                size="large" 
                style={{ color: '#1890ff', marginBottom: '8px' }} 
              />
              <Title heading={3} style={{ margin: '8px 0' }}>
                {blacklistStats.total_blacklisted || 0}
              </Title>
              <Text type="secondary">{t('黑名单IP')}</Text>
            </div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <IconMonitor 
                size="large" 
                style={{ color: '#52c41a', marginBottom: '8px' }} 
              />
              <Title heading={3} style={{ margin: '8px 0' }}>
                {streamStats.active_connections || 0}
              </Title>
              <Text type="secondary">{t('活跃流连接')}</Text>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 详细信息 */}
      <Row gutter={24}>
        <Col span={12}>
          <Card title={t('防护配置状态')}>
            <Descriptions data={[
              {
                key: t('异常检测'),
                value: (
                  <Tag color={config.abnormal_detection?.enabled ? 'green' : 'red'}>
                    {config.abnormal_detection?.enabled ? t('已启用') : t('已禁用')}
                  </Tag>
                )
              },
              {
                key: t('请求大小限制'),
                value: (
                  <Tag color={config.request_size_limit?.enabled ? 'green' : 'red'}>
                    {config.request_size_limit?.enabled ? t('已启用') : t('已禁用')}
                  </Tag>
                )
              },
              {
                key: t('流保护'),
                value: (
                  <Tag color={config.stream_protection?.enabled ? 'green' : 'red'}>
                    {config.stream_protection?.enabled ? t('已启用') : t('已禁用')}
                  </Tag>
                )
              },
              {
                key: t('IP黑名单'),
                value: (
                  <Tag color={config.ip_blacklist?.enabled ? 'green' : 'red'}>
                    {config.ip_blacklist?.enabled ? t('已启用') : t('已禁用')}
                  </Tag>
                )
              }
            ]} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title={t('可疑活动统计')}>
            {securityStats.suspicious_activities && Object.keys(securityStats.suspicious_activities).length > 0 ? (
              <Descriptions data={
                Object.entries(securityStats.suspicious_activities).map(([key, value]) => ({
                  key: key,
                  value: value
                }))
              } />
            ) : (
              <Text type="secondary">{t('暂无可疑活动')}</Text>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default SecurityOverview;
