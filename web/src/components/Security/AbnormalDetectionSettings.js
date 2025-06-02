import React, { useState, useEffect } from 'react';
import { Card, Form, Row, Col, Button, Switch, InputNumber, Slider, Typography, Space, Divider, Spin } from '@douyinfe/semi-ui';
import { IconAlertTriangle, IconSave } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess, showWarning } from '../../helpers';

const { Title, Text } = Typography;

const AbnormalDetectionSettings = ({ config, refresh }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    enabled: true,
    max_prompt_length: 50000,
    max_random_char_ratio: 0.8,
    min_request_interval: 100,
    suspicious_score_limit: 100,
    max_concurrent_streams: 5,
    stream_timeout_seconds: 300,
    auto_blacklist_enabled: true
  });

  useEffect(() => {
    if (config?.abnormal_detection) {
      setFormData({
        enabled: config.abnormal_detection.enabled || false,
        max_prompt_length: config.abnormal_detection.max_prompt_length || 50000,
        max_random_char_ratio: config.abnormal_detection.max_random_char_ratio || 0.8,
        min_request_interval: config.abnormal_detection.min_request_interval_ms || 100,
        suspicious_score_limit: config.abnormal_detection.suspicious_score_limit || 100,
        max_concurrent_streams: config.abnormal_detection.max_concurrent_streams || 5,
        stream_timeout_seconds: config.abnormal_detection.stream_timeout_seconds || 300,
        auto_blacklist_enabled: config.abnormal_detection.auto_blacklist_enabled || true
      });
    }
  }, [config]);

  const handleSave = async () => {
    setLoading(true);
    try {
      const newConfig = {
        ...(config || {}),
        abnormal_detection: {
          enabled: formData.enabled,
          max_prompt_length: formData.max_prompt_length,
          max_random_char_ratio: formData.max_random_char_ratio,
          min_request_interval_ms: formData.min_request_interval,
          suspicious_score_limit: formData.suspicious_score_limit,
          max_concurrent_streams: formData.max_concurrent_streams,
          stream_timeout_seconds: formData.stream_timeout_seconds,
          auto_blacklist_enabled: formData.auto_blacklist_enabled
        }
      };

      const response = await API.put('/api/security/config', newConfig);
      
      if (response.data.success) {
        showSuccess(t('异常检测设置保存成功'));
        refresh();
      } else {
        showError(response.data.message || t('保存失败'));
      }
    } catch (error) {
      showError(t('保存异常检测设置失败'));
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

  if (!config) {
    return <Spin spinning />;
  }

  return (
    <div>
      <Card>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <IconAlertTriangle size="large" style={{ color: '#fa8c16' }} />
            <Title heading={4} style={{ margin: 0 }}>
              {t('异常行为检测配置')}
            </Title>
          </Space>
          <Text type="secondary" style={{ display: 'block', marginTop: '8px' }}>
            {t('检测和阻止恶意请求，如token浪费攻击、高频请求等')}
          </Text>
        </div>

        <Form layout="vertical">
          {/* 基础设置 */}
          <Card title={t('基础设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('启用异常检测')}>
                  <Switch
                    checked={formData.enabled}
                    onChange={(checked) => handleFieldChange('enabled', checked)}
                    checkedText={t('开启')}
                    uncheckedText={t('关闭')}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('总开关，关闭后所有异常检测功能将停用')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('自动加入黑名单')}>
                  <Switch
                    checked={formData.auto_blacklist_enabled}
                    onChange={(checked) => handleFieldChange('auto_blacklist_enabled', checked)}
                    checkedText={t('开启')}
                    uncheckedText={t('关闭')}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('检测到恶意行为时自动将IP加入黑名单')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 内容检测设置 */}
          <Card title={t('内容检测设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('最大Prompt长度')}>
                  <InputNumber
                    value={formData.max_prompt_length}
                    onChange={(value) => handleFieldChange('max_prompt_length', value)}
                    min={1000}
                    max={200000}
                    step={1000}
                    suffix={t('字符')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('超过此长度的Prompt将被视为可疑')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('随机字符比例阈值')}>
                  <div>
                    <Slider
                      value={formData.max_random_char_ratio}
                      onChange={(value) => handleFieldChange('max_random_char_ratio', value)}
                      min={0.1}
                      max={1.0}
                      step={0.1}
                      marks={{
                        0.1: '10%',
                        0.5: '50%',
                        0.8: '80%',
                        1.0: '100%'
                      }}
                    />
                    <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                      {t('当前值')}: {Math.round(formData.max_random_char_ratio * 100)}%
                    </Text>
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 频率检测设置 */}
          <Card title={t('频率检测设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('最小请求间隔')}>
                  <InputNumber
                    value={formData.min_request_interval}
                    onChange={(value) => handleFieldChange('min_request_interval', value)}
                    min={10}
                    max={5000}
                    step={10}
                    suffix={t('毫秒')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('低于此间隔的请求将被视为高频攻击')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('可疑分数限制')}>
                  <InputNumber
                    value={formData.suspicious_score_limit}
                    onChange={(value) => handleFieldChange('suspicious_score_limit', value)}
                    min={10}
                    max={500}
                    step={10}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('超过此分数的IP将被临时封禁')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 流检测设置 */}
          <Card title={t('流检测设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('最大并发流数量')}>
                  <InputNumber
                    value={formData.max_concurrent_streams}
                    onChange={(value) => handleFieldChange('max_concurrent_streams', value)}
                    min={1}
                    max={20}
                    step={1}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('单个IP允许的最大并发流请求数')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('流超时时间')}>
                  <InputNumber
                    value={formData.stream_timeout_seconds}
                    onChange={(value) => handleFieldChange('stream_timeout_seconds', value)}
                    min={30}
                    max={1800}
                    step={30}
                    suffix={t('秒')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('流连接的最大持续时间')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
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

export default AbnormalDetectionSettings;
