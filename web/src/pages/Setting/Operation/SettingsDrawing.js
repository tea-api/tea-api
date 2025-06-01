import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin, Tag } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsDrawing(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    DrawingEnabled: false,
    MjNotifyEnabled: false,
    MjAccountFilterEnabled: false,
    MjForwardUrlEnabled: false,
    MjModeClearEnabled: false,
    MjActionCheckSuccessEnabled: false,
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
    // 确保 props.options 存在且为对象
    if (!props.options || typeof props.options !== 'object') {
      console.warn('props.options is invalid:', props.options);
      return;
    }

    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }

    // 确保 currentInputs 不为空，合并默认值
    const mergedInputs = { ...inputs, ...currentInputs };
    setInputs(prevInputs => ({ ...prevInputs, ...currentInputs }));
    setInputsRow(structuredClone(mergedInputs));
    if (refForm.current) refForm.current.setValues(mergedInputs);
    localStorage.setItem('mj_notify_enabled', String(mergedInputs.MjNotifyEnabled || false));
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('绘图设置')}>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'DrawingEnabled'}
                  label={t('启用绘图功能')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) => {
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      DrawingEnabled: value,
                    }));
                  }}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'MjNotifyEnabled'}
                  label={t('允许回调（会泄露服务器 IP 地址）')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) =>
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      MjNotifyEnabled: value,
                    }))
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'MjAccountFilterEnabled'}
                  label={t('允许 AccountFilter 参数')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) =>
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      MjAccountFilterEnabled: value,
                    }))
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'MjForwardUrlEnabled'}
                  label={t('开启之后将上游地址替换为服务器地址')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) =>
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      MjForwardUrlEnabled: value,
                    }))
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'MjModeClearEnabled'}
                  label={
                    <>
                      {t('开启之后会清除用户提示词中的')} <Tag>--fast</Tag> 、
                      <Tag>--relax</Tag> {t('以及')} <Tag>--turbo</Tag>{' '}
                      {t('参数')}
                    </>
                  }
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) =>
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      MjModeClearEnabled: value,
                    }))
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'MjActionCheckSuccessEnabled'}
                  label={t('检测必须等待绘图成功才能进行放大等操作')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) =>
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      MjActionCheckSuccessEnabled: value,
                    }))
                  }
                />
              </Col>
            </Row>
            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存绘图设置')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
