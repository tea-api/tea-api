import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin, Switch, Space, Typography } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';

export default function LinuxDOTrustLevelSettings(props) {
  const { t } = useTranslation();
  const { Text } = Typography;
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    'linuxdo_trust_level.enabled': 'false',
    'linuxdo_trust_level.trust_level_0': '0',
    'linuxdo_trust_level.trust_level_1': '1000',
    'linuxdo_trust_level.trust_level_2': '2000',
    'linuxdo_trust_level.trust_level_3': '5000',
    'linuxdo_trust_level.trust_level_4': '10000',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof inputs[item.key] === 'boolean') {
        value = String(inputs[item.key]);
      } else {
        value = inputs[item.key];
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('L站信任等级额度设置')}>
            <Row gutter={16}>
              <Col xs={24}>
                <Space>
                  <Switch 
                    checked={inputs['linuxdo_trust_level.enabled'] === 'true'}
                    onChange={(v) => {
                      setInputs({
                        ...inputs,
                        'linuxdo_trust_level.enabled': v ? 'true' : 'false',
                      });
                    }}
                  />
                  <Text>{t('启用L站信任等级额度赠送')}</Text>
                </Space>
              </Col>
            </Row>

            <Row gutter={16} style={{ marginTop: 16 }}>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('信任等级 0 额度')}
                  field={'linuxdo_trust_level.trust_level_0'}
                  step={100}
                  min={0}
                  suffix={'Token'}
                  placeholder={''}
                  disabled={inputs['linuxdo_trust_level.enabled'] !== 'true'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'linuxdo_trust_level.trust_level_0': String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('信任等级 1 额度')}
                  field={'linuxdo_trust_level.trust_level_1'}
                  step={100}
                  min={0}
                  suffix={'Token'}
                  placeholder={''}
                  disabled={inputs['linuxdo_trust_level.enabled'] !== 'true'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'linuxdo_trust_level.trust_level_1': String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('信任等级 2 额度')}
                  field={'linuxdo_trust_level.trust_level_2'}
                  step={100}
                  min={0}
                  suffix={'Token'}
                  placeholder={''}
                  disabled={inputs['linuxdo_trust_level.enabled'] !== 'true'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'linuxdo_trust_level.trust_level_2': String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('信任等级 3 额度')}
                  field={'linuxdo_trust_level.trust_level_3'}
                  step={100}
                  min={0}
                  suffix={'Token'}
                  placeholder={''}
                  disabled={inputs['linuxdo_trust_level.enabled'] !== 'true'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'linuxdo_trust_level.trust_level_3': String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('信任等级 4 额度')}
                  field={'linuxdo_trust_level.trust_level_4'}
                  step={100}
                  min={0}
                  suffix={'Token'}
                  placeholder={''}
                  disabled={inputs['linuxdo_trust_level.enabled'] !== 'true'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'linuxdo_trust_level.trust_level_4': String(value),
                    })
                  }
                />
              </Col>
            </Row>

            <Row style={{ marginTop: 16 }}>
              <Button size='default' onClick={onSubmit} disabled={loading}>
                {t('保存设置')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
} 