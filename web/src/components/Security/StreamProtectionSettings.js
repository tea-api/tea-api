import React, { useState, useEffect } from 'react';
import { Card, Form, Row, Col, Button, Switch, InputNumber, Typography, Space, Divider } from '@douyinfe/semi-ui';
import { IconShield, IconSave } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';

const { Title, Text } = Typography;

const StreamProtectionSettings = ({ config, refresh }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    enabled: true,
    max_streams_per_ip: 3,
    max_streams_per_user: 5,
    stream_idle_timeout_sec: 30,
    stream_max_duration_sec: 600,
    min_bytes_per_second: 10,
    slow_client_timeout_sec: 60
  });

  useEffect(() => {
    if (config?.stream_protection) {
      setFormData({
        enabled: config.stream_protection.enabled || false,
        max_streams_per_ip: config.stream_protection.max_streams_per_ip || 3,
        max_streams_per_user: config.stream_protection.max_streams_per_user || 5,
        stream_idle_timeout_sec: config.stream_protection.stream_idle_timeout_sec || 30,
        stream_max_duration_sec: config.stream_protection.stream_max_duration_sec || 600,
        min_bytes_per_second: config.stream_protection.min_bytes_per_second || 10,
        slow_client_timeout_sec: config.stream_protection.slow_client_timeout_sec || 60
      });
    }
  }, [config]);

  const handleSave = async () => {
    setLoading(true);
    try {
      const newConfig = {
        ...config,
        stream_protection: {
          enabled: formData.enabled,
          max_streams_per_ip: formData.max_streams_per_ip,
          max_streams_per_user: formData.max_streams_per_user,
          stream_idle_timeout_sec: formData.stream_idle_timeout_sec,
          stream_max_duration_sec: formData.stream_max_duration_sec,
          min_bytes_per_second: formData.min_bytes_per_second,
          slow_client_timeout_sec: formData.slow_client_timeout_sec
        }
      };

      const response = await API.put('/api/security/config', newConfig);
      
      if (response.data.success) {
        showSuccess(t('流保护设置保存成功'));
        refresh();
      } else {
        showError(response.data.message || t('保存失败'));
      }
    } catch (error) {
      showError(t('保存流保护设置失败'));
    } finally {
      setLoading(false);
    }
  };

  const handleFieldChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const formatDuration = (seconds) => {
    if (seconds < 60) return `${seconds}秒`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
    return `${Math.floor(seconds / 3600)}小时${Math.floor((seconds % 3600) / 60)}分`;
  };

  return (
    <div>
      <Card>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <IconShield size="large" style={{ color: '#52c41a' }} />
            <Title heading={4} style={{ margin: 0 }}>
              {t('流保护配置')}
            </Title>
          </Space>
          <Text type="secondary" style={{ display: 'block', marginTop: '8px' }}>
            {t('保护流式请求，防止慢速读取攻击和资源滥用')}
          </Text>
        </div>

        <Form layout="vertical">
          {/* 基础设置 */}
          <Card title={t('基础设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={24}>
                <Form.Item label={t('启用流保护')}>
                  <Switch
                    checked={formData.enabled}
                    onChange={(checked) => handleFieldChange('enabled', checked)}
                    checkedText={t('开启')}
                    uncheckedText={t('关闭')}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('总开关，关闭后所有流保护功能将停用')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 并发限制设置 */}
          <Card title={t('并发限制设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('每个IP最大流数量')}>
                  <InputNumber
                    value={formData.max_streams_per_ip}
                    onChange={(value) => handleFieldChange('max_streams_per_ip', value)}
                    min={1}
                    max={20}
                    step={1}
                    suffix={t('个')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('单个IP地址允许的最大并发流连接数')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('每个用户最大流数量')}>
                  <InputNumber
                    value={formData.max_streams_per_user}
                    onChange={(value) => handleFieldChange('max_streams_per_user', value)}
                    min={1}
                    max={50}
                    step={1}
                    suffix={t('个')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('单个用户允许的最大并发流连接数')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 超时设置 */}
          <Card title={t('超时设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('流空闲超时')}>
                  <InputNumber
                    value={formData.stream_idle_timeout_sec}
                    onChange={(value) => handleFieldChange('stream_idle_timeout_sec', value)}
                    min={5}
                    max={300}
                    step={5}
                    suffix={t('秒')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('流连接无数据传输的最大空闲时间')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('流最大持续时间')}>
                  <InputNumber
                    value={formData.stream_max_duration_sec}
                    onChange={(value) => handleFieldChange('stream_max_duration_sec', value)}
                    min={60}
                    max={3600}
                    step={60}
                    suffix={t('秒')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('单个流连接的最大持续时间')} ({formatDuration(formData.stream_max_duration_sec)})
                  </Text>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('慢客户端超时')}>
                  <InputNumber
                    value={formData.slow_client_timeout_sec}
                    onChange={(value) => handleFieldChange('slow_client_timeout_sec', value)}
                    min={10}
                    max={300}
                    step={10}
                    suffix={t('秒')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('检测到慢客户端后的超时时间')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('最小传输速率')}>
                  <InputNumber
                    value={formData.min_bytes_per_second}
                    onChange={(value) => handleFieldChange('min_bytes_per_second', value)}
                    min={1}
                    max={1000}
                    step={1}
                    suffix={t('字节/秒')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('低于此速率的连接将被视为慢客户端')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 配置预览 */}
          <Card title={t('配置预览')} style={{ marginBottom: '24px' }}>
            <div style={{ background: '#f6f8fa', padding: '16px', borderRadius: '6px' }}>
              <Text type="secondary">
                {t('当前配置将允许：')}
              </Text>
              <ul style={{ marginTop: '8px', marginBottom: 0 }}>
                <li>{t('每个IP最多')} {formData.max_streams_per_ip} {t('个并发流连接')}</li>
                <li>{t('每个用户最多')} {formData.max_streams_per_user} {t('个并发流连接')}</li>
                <li>{t('流连接最长持续')} {formatDuration(formData.stream_max_duration_sec)}</li>
                <li>{t('空闲超时')} {formData.stream_idle_timeout_sec} {t('秒')}</li>
                <li>{t('最小传输速率')} {formData.min_bytes_per_second} {t('字节/秒')}</li>
              </ul>
            </div>
          </Card>

          <Divider />
          
          <div style={{ textAlign: 'right' }}>
            <Button
              type="primary"
              icon={<IconSave />}
              loading={loading}
              onClick={handleSave}
              size="large"
            >
              {t('保存设置')}
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default StreamProtectionSettings;
