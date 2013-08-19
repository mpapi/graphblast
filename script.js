/*global d3:false, EventSource:false */
(function () {
  'use strict';

  // TODO: it might be better to rotate the whole thing, and then individually
  // rotate each piece of text back

  // Returns a value for the SVG "transform" property for translating by x, y.
  var _translate = function (x, y) {
    return ['translate(', x, ',', y, ')'].join('');
  };

  var setter = function (params, func) {
    if (func === undefined) {
      func = 'attr';
    }
    return function (node) {
      d3.entries(params).forEach(function (item) {
        node[func](item.key, item.value);
      });
      return node;
    };
  };

  var Orientation = {
    wide: function (data, axisLength, barLength) {
      return {
        axis: {orient: 'left', transform: ''},
        range: {
          x: [0, axisLength - axisLength / data.length],
          y: [0, barLength]
        },
        svg: {width: barLength + 105, height: axisLength + 65},
        label: {
          x: -35,
          y: (axisLength - axisLength / data.length) * 0.5,
          rotate: -90
        },
        translate: function (x) {
          return function (d) {
            return _translate(0, x(d.x));
          };
        },
        bar: function (x, y, dx) {
          return setter({
            x: 0,
            y: 1,
            height: Math.max(1, dx - 1),
            width: function (d) { return y(d.y); }
          });
        },
        text: function (x, y, dx) {
          return setter({
            x: function (d) {
              var len = this.getComputedTextLength();
              return y(d.y) > len + 30 ? y(d.y) - 6 - len : y(d.y) + 6;
            },
            y: dx * 0.5,
            'class': function (d) {
              var len = this.getComputedTextLength();
              return y(d.y) > len + 30 ? 'inside' : 'outside';
            },
            'dominant-baseline': 'middle'
          });
        }
      };
    },

    tall: function (data, axisLength, barLength) {
      return {
        axis: {orient: 'bottom', transform: _translate(0, barLength)},
        range: {
          x: [0, axisLength - axisLength / data.length],
          y: [barLength, 0]
        },
        svg: {width: axisLength + 65, height: barLength + 105},
        label: {
          x: (axisLength - axisLength / data.length) * 0.5,
          y: barLength + 50,
          rotate: 0
        },
        translate: function (x, y) {
          return function (d) {
            return _translate(x(d.x), y(d.y));
          };
        },
        bar: function (x, y, dx) {
          return setter({
            x: 1,
            y: 0,
            height: function (d) { return barLength - y(d.y); },
            width: Math.max(1, dx - 1)
          });
        },
        text: function (x, y, dx) {
          return setter({
            x: dx * 0.5,
            y: function (d) {
              return y(d.y) > barLength - 30 ? -6 : 18;
            },
            'class': function (d) {
              return y(d.y) > barLength - 30 ? 'outside' : 'inside';
            },
            'text-anchor': 'middle'
          });
        }
      };
    }
  };

  var orientation = function (opts, data) {
    var axisLength = opts.Wide ? opts.Height : opts.Width;
    var barLength = opts.Wide ? opts.Width : opts.Height;
    var orientation = opts.Wide ? 'wide' : 'tall';
    return Orientation[orientation](data, axisLength, barLength);
  };

  var parseColors = function (colors) {
    var parts = colors ? colors.split(',') : [];
    return {
      bg: parts[0],
      fg: parts[1],
      bar: parts[2]
    };
  };

  var CSS = {
    colors: function (colors) {
      var styles = [];
      if (colors.bg && colors.fg) {
        styles.push('body { background-color: ' + colors.bg + '}');
        styles.push('.axis path, .axis line { stroke: ' + colors.fg + '}');
        styles.push('text, text.outside { fill: ' + colors.fg + '}');
        styles.push('text.inside { fill: ' + colors.bg + '}');
      }
      if (colors.bar) {
        styles.push('.bar { fill: ' + colors.bar + '}');
      }
      return styles.join('\n');
    },
    overrides: function () {
      var overrides = d3.select('style.overrides');
      if (overrides.empty()) {
        overrides = d3.select('head').append('style')
          .classed('overrides', true);
      }
      return overrides;
    }
  };

  var applyStyle = function (opts) {
    if (opts.Label) {
      document.title = opts.Label;
    }
    var styles = [CSS.colors(parseColors(opts.Colors))];
    if (opts.FontSize) {
      styles.push('body { font-size: ' + opts.FontSize + '}');
    }
    CSS.overrides().text(styles.join('\n'));
  };

  var histogram = function (data, opts) {

    if (data.length <= 1) {
      // TODO: don't return, show something
      return;
    }

    applyStyle(opts);

    // TODO: dynamic with screen resize, underscore debounce?
    var orient = orientation(opts, data);

    var x = d3.scale.linear()
      .domain([d3.min(data, function (d) { return d.x; }),
               d3.max(data, function (d) { return d.x; }) + opts.Bucket])
      .range(orient.range.x);

    var y = d3.scale.linear()
      .domain([0, d3.max(data, function (d) { return d.y; })])
      .range(orient.range.y);

    var dx = (x(1) - x(0)) * opts.Bucket;

    var axis = d3.svg.axis().scale(x).orient(orient.axis.orient);

    var svg = d3.select('body').append('svg')
      .attr('width', orient.svg.width)
      .attr('height', orient.svg.height)
      .append('g')
      .attr('transform', _translate(50, 50));
      // TODO translate amount based on axis width & svg width

    svg.append('g')
      .attr('transform', _translate(orient.label.x, orient.label.y))
      .append('text')
      .text(opts.Label)
      .attr('class', 'label')
      .attr('text-anchor', 'middle')
      .attr('font-size', '1.1em')
      .attr('font-weight', 'bold')
      .attr('transform', 'rotate(' + orient.label.rotate + ')');

    var bar = svg.selectAll('.bar').data(data)
      .enter()
      .append('g')
      .attr('class', 'bar')
      .attr('transform', orient.translate(x, y, dx));

    orient.bar(x, y, dx)(
      bar.append('rect'));
    orient.text(x, y, dx)(
      bar.append('text').text(function (d) { return d.y; }));

    svg.append('g')
      .attr('class', 'axis')
      .attr('transform', orient.axis.transform)
      .call(axis);
  };

  var pushHistogram = window.pushHistogram = function (data) {
    var hist = d3.map(data.Values).entries().map(function (i) {
      return {x: parseFloat(i.key), y: i.value};
    });
    hist.sort(d3.ascending);
    d3.select('svg').remove();
    histogram(hist, data);
  };

  var timeSeries = function (data, opts) {
    if (data.length <= 1) {
      // TODO: don't return, show something
      return;
    }

    //applyStyle(opts);
    var width = opts.Width; // + 65;
    var height = opts.Height; // + 200;

    var x = d3.time.scale()
      .domain(d3.extent(data, function (d) { return d.x; }))
      .range([0, width]);

    var y = d3.scale.linear()
      .domain(d3.extent(data, function (d) { return d.y; }))
      .range([height, 0]);

    var xAxis = d3.svg.axis().scale(x).orient('bottom');
    var yAxis = d3.svg.axis().scale(y).orient('left');
    var line = d3.svg.line()
      .x(function (d) { return x(d.x); })
      .y(function (d) { return y(d.y); });

    var svg = d3.select('body').append('svg')
      .attr('width', width + 65)
      .attr('height', height + 105)
      .append('g')
      .attr('transform', _translate(50, 50));
      // TODO translate amount based on axis width & svg width

    svg.append('g')
      .attr('transform', _translate(width * 0.5, height + 50))
      .append('text')
      .text(opts.Label)
      .attr('class', 'label')
      .attr('text-anchor', 'middle')
      .attr('font-size', '1.1em')
      .attr('font-weight', 'bold');
      //.attr('transform', 'rotate(' + orient.label.rotate + ')');


    svg.append('path')
      .datum(data)
      .attr('class', 'line')
      .attr('d', line);

    svg.append('g')
      .attr('class', 'y axis')
      .call(yAxis);

    svg.append('g')
      .attr('class', 'x axis')
      .attr('transform', _translate(0, y(Math.max(0, y.domain()[0]))))
      .call(xAxis);
  };

  var pushTimeSeries = window.pushTimeSeries = function (data) {
    var ts = d3.map(data.Values).entries().sort(function (a, b) {
      return d3.ascending(a.key, b.key);
    }).map(function (i) {
      return {x: new Date(i.key), y: i.value};
    });
    d3.select('svg').remove();
    timeSeries(ts, data);
  };

  var events = new EventSource('/data');
  events.onmessage = function (e) {
    var data = JSON.parse(e.data);
    if (data.type && data.type === 'error') {
      // TODO show the error
      console.error(data);
      return;
    }
    console.debug(data);
    // TODO show EOF, others
    // TODO switch on "histogram" type
    if (data.Layout === 'histogram') {
      pushHistogram(data);
    } else if (data.Layout == 'time-series') {
      pushTimeSeries(data);
    }
  };
})();
