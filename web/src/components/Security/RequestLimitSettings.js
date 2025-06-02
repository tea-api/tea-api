import React, { useState, useEffect } from 'react';
import { Card, Form, Row, Col, Button, Switch, InputNumber, Typography, Space, Divider, Spin } from '@douyinfe/semi-ui';
import { IconSetting, IconSave } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';

const { Title, Text } = Typography;

const RequestLimitSettings = ({ config, refresh }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    enabled: true,
    max_request_body_size: 10485760, // 10MB
    max_prompt_length: 100000,
    max_messages_count: 100,
    max_single_message_size: 50000,
    max_tokens_limit: 100000,
    content_validation: true
  });

  useEffect(() => {
    if (config?.request_size_limit) {
      setFormData({
        enabled: config.request_size_limit.enabled || false,
        max_request_body_size: config.request_size_limit.max_request_body_size || 10485760,
        max_prompt_length: config.request_size_limit.max_prompt_length || 100000,
        max_messages_count: config.request_size_limit.max_messages_count || 100,
        max_single_message_size: config.request_size_limit.max_single_message_size || 50000,
        max_tokens_limit: config.request_size_limit.max_tokens_limit || 100000,
        content_validation: config.request_size_limit.content_validation !== false
      });
    }
  }, [config]);

  const handleSave = async () => {
    setLoading(true);
    try {
      const newConfig = {
        ...(config || {}),
        request_size_limit: {
          enabled: formData.enabled,
          max_request_body_size: formData.max_request_body_size,
          max_prompt_length: formData.max_prompt_length,
          max_messages_count: formData.max_messages_count,
          max_single_message_size: formData.max_single_message_size,
          max_tokens_limit: formData.max_tokens_limit,
          content_validation: formData.content_validation
        }
      };

      const response = await API.put('/api/security/config', newConfig);
      
      if (response.data.success) {
        showSuccess(t('请求限制设置保存成功'));
        refresh();
      } else {
        showError(response.data.message || t('保存失败'));
      }
    } catch (error) {
      showError(t('保存请求限制设置失败'));
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

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div>
      <Card>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <IconSetting size="large" style={{ color: '#1890ff' }} />
            <Title heading={4} style={{ margin: 0 }}>
              {t('请求大小限制配置')}
            </Title>
          </Space>
          <Text type="secondary" style={{ display: 'block', marginTop: '8px' }}>
            {t('限制请求体大小和内容，防止超大请求攻击')}
          </Text>
        </div>

        <Form layout="vertical">
          {/* 基础设置 */}
          <Card title={t('基础设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('启用请求大小限制')}>
                  <Switch
                    checked={formData.enabled}
                    onChange={(checked) => handleFieldChange('enabled', checked)}
                    checkedText={t('开启')}
                    uncheckedText={t('关闭')}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('总开关，关闭后所有请求大小限制将停用')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('内容质量验证')}>
                  <Switch
                    checked={formData.content_validation}
                    onChange={(checked) => handleFieldChange('content_validation', checked)}
                    checkedText={t('开启')}
                    uncheckedText={t('关闭')}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('检测随机内容、重复内容等异常模式')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 请求体大小限制 */}
          <Card title={t('请求体大小限制')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={24}>
                <Form.Item label={t('最大请求体大小')}>
                  <div>
                    <InputNumber
                      value={formData.max_request_body_size}
                      onChange={(value) => handleFieldChange('max_request_body_size', value)}
                      min={1048576} // 1MB
                      max={104857600} // 100MB
                      step={1048576}
                      style={{ width: '300px' }}
                      formatter={(value) => formatBytes(value)}
                      parser={(value) => {
                        const match = value.match(/^([\d.]+)\s*([KMGT]?B)$/i);
                        if (!match) return value;
                        const [, num, unit] = match;
                        const multipliers = { B: 1, KB: 1024, MB: 1024*1024, GB: 1024*1024*1024 };
                        return Math.round(parseFloat(num) * (multipliers[unit.toUpperCase()] || 1));
                      }}
                    />
                    <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                      {t('当前值')}: {formatBytes(formData.max_request_body_size)}
                    </Text>
                  </div>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 内容限制设置 */}
          <Card title={t('内容限制设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('最大Prompt总长度')}>
                  <InputNumber
                    value={formData.max_prompt_length}
                    onChange={(value) => handleFieldChange('max_prompt_length', value)}
                    min={1000}
                    max={500000}
                    step={1000}
                    suffix={t('字符')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('所有消息内容的总长度限制')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('单条消息最大长度')}>
                  <InputNumber
                    value={formData.max_single_message_size}
                    onChange={(value) => handleFieldChange('max_single_message_size', value)}
                    min={1000}
                    max={200000}
                    step={1000}
                    suffix={t('字符')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('单条消息的最大长度限制')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('最大消息数量')}>
                  <InputNumber
                    value={formData.max_messages_count}
                    onChange={(value) => handleFieldChange('max_messages_count', value)}
                    min={1}
                    max={1000}
                    step={1}
                    suffix={t('条')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('单次请求中的最大消息数量')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('最大tokens限制')}>
                  <InputNumber
                    value={formData.max_tokens_limit}
                    onChange={(value) => handleFieldChange('max_tokens_limit', value)}
                    min={1000}
                    max={1000000}
                    step={1000}
                    suffix={t('tokens')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('请求中max_tokens参数的最大值')}
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

export default RequestLimitSettings;
