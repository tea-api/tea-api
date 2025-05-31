import React, { useEffect, useRef, useState } from 'react';
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsAbnormalDetection(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    enabled: false,
    hf_enabled: false,
    max_requests_per_second: 50,
    sleep_seconds: 0,
  });
  const [inputsRow, setInputsRow] = useState(inputs);
  const [rulesObj, setRulesObj] = useState({});
  const [securityObj, setSecurityObj] = useState({});
  const refForm = useRef();

  function onSubmit() {
    const changed = compareObjects(inputsRow, inputs);
    if (!changed.length) return showWarning(t('你似乎并没有修改什么'));

    const requestQueue = [];
    if (inputs.enabled !== inputsRow.enabled) {
      requestQueue.push(
        API.put('/api/option/', {
          key: 'abnormal_detection.enabled',
          value: String(inputs.enabled),
        }),
      );
    }
    if (
      inputs.hf_enabled !== inputsRow.hf_enabled ||
      inputs.max_requests_per_second !== inputsRow.max_requests_per_second
    ) {
      const newRules = {
        ...rulesObj,
        high_frequency: {
          enabled: inputs.hf_enabled,
          max_requests_per_second: Number(inputs.max_requests_per_second),
        },
      };
      requestQueue.push(
        API.put('/api/option/', {
          key: 'abnormal_detection.rules',
          value: JSON.stringify(newRules),
        }),
      );
    }
    if (inputs.sleep_seconds !== inputsRow.sleep_seconds) {
      const newSecurity = {
        ...securityObj,
        sleep_seconds: Number(inputs.sleep_seconds),
      };
      requestQueue.push(
        API.put('/api/option/', {
          key: 'abnormal_detection.security',
          value: JSON.stringify(newSecurity),
        }),
      );
    }

    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1 && res.includes(undefined)) return;
        if (requestQueue.length > 1 && res.includes(undefined))
          return showError(t('部分保存失败，请重试'));
        for (let r of res) {
          if (!r.data.success) return showError(r.data.message);
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
    const currentInputs = { ...inputs };
    const opt = props.options || {};
    if (opt['abnormal_detection.enabled'] !== undefined) {
      currentInputs.enabled =
        opt['abnormal_detection.enabled'] === true ||
        opt['abnormal_detection.enabled'] === 'true';
    }
    let rules = {
      high_frequency: { enabled: false, max_requests_per_second: 50 },
    };
    if (opt['abnormal_detection.rules']) {
      try {
        rules = JSON.parse(opt['abnormal_detection.rules']);
      } catch (e) {}
    }
    let security = { sleep_seconds: 0 };
    if (opt['abnormal_detection.security']) {
      try {
        security = JSON.parse(opt['abnormal_detection.security']);
      } catch (e) {}
    }
    setRulesObj(rules);
    setSecurityObj(security);
    currentInputs.hf_enabled = rules.high_frequency?.enabled || false;
    currentInputs.max_requests_per_second =
      rules.high_frequency?.max_requests_per_second || 50;
    currentInputs.sleep_seconds = security.sleep_seconds || 0;

    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    if (refForm.current) refForm.current.setValues(currentInputs);
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('异常行为检测')}>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'enabled'}
                  label={t('启用异常行为检测')}
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(val) => setInputs({ ...inputs, enabled: val })}
                />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'hf_enabled'}
                  label={t('启用高频调用检测')}
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(val) => setInputs({ ...inputs, hf_enabled: val })}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  field={'max_requests_per_second'}
                  label={t('单位时间最大请求数')}
                  step={1}
                  min={1}
                  onChange={(val) =>
                    setInputs({ ...inputs, max_requests_per_second: val })
                  }
                />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  field={'sleep_seconds'}
                  label={t('安全延迟')}
                  suffix={t('秒')}
                  step={1}
                  min={0}
                  onChange={(val) =>
                    setInputs({ ...inputs, sleep_seconds: val })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Button onClick={onSubmit}>{t('保存异常检测设置')}</Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
